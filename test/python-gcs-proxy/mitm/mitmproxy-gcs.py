# Run this proxy by running the below command:
# > mitmproxy -s mitmproxy-gcs.py
import base64
from email import parser
import hashlib
import json
import logging
import os
import re

from mitmproxy import http
from mitmproxy.net.http import status_codes
import tink
from tink import aead
from tink.integration import gcpkms
import debugpy
import time

# Env variables for GCP KMS with Key encryption keys. The default values are set for testing locally.
GCP_KMS_PROJECT_ID = os.environ.get("GCP_KMS_PROJECT_ID", "mando-host-project")
GCP_KMS_KEY = os.environ.get(
    "GCP_KMS_KEY",
    "gcp-kms://projects/mando-host-project/locations/global/keyRings/test/cryptoKeys/proxy-kek",
)
GCP_KMS_CREDENTIALS = os.environ.get("GCP_KMS_CREDENTIALS", None)
GCS_PROXY_DISABLE_ENCRYPTION = os.environ.get("GCS_PROXY_DISABLE_ENCRYPTION", "false").lower() == "true"



LOG_LEVEL_STR = os.environ.get("GCS_PROXY_LOG_LEVEL", "INFO")
log_level = getattr(logging, LOG_LEVEL_STR.upper(), logging.INFO)
logger = logging.getLogger("gcs-proxy")
logger.setLevel(log_level)

handler = logging.StreamHandler()
handler.setLevel(log_level)

formatter = logging.Formatter(
    "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
handler.setFormatter(formatter)

logger.addHandler(handler)

if GCS_PROXY_DISABLE_ENCRYPTION:
  logger.info(f"GCS_PROXY_DISABLE_ENCRYPTION is set to {GCS_PROXY_DISABLE_ENCRYPTION}. Encryption is disabled.")
else :
  logger.info(f"GCS_PROXY_DISABLE_ENCRYPTION is set to {GCS_PROXY_DISABLE_ENCRYPTION}. Encryption is enabled.")

# Intialize kms client, tink and create envelope AEAD primitive using AES256 GCM for encrypting the data
aead.register()

try:
  client = gcpkms.GcpKmsClient(GCP_KMS_KEY, GCP_KMS_CREDENTIALS)
except tink.TinkError as e:
  logger.exception("Error creating GCP KMS client: %s", e)
  raise  # Re-raise the exception to be handled at a higher level

try:
  remote_aead = client.get_aead(GCP_KMS_KEY)
  env_aead = aead.KmsEnvelopeAead(
      aead.aead_key_templates.AES256_GCM, remote_aead
  )
except tink.TinkError as e:
  logger.exception("Error creating primitive: %s", e)
  raise


def _base64_md5hash(buffer_object: bytes) -> str:
  """Get MD5 hash of bytes (as base64).

  Args:
      buffer_object: Buffer containing bytes used to compute an MD5 hash.

  Returns:
      A base64 encoded digest of the MD5 hash.
  """
  hash_obj = hashlib.md5(buffer_object)
  digest_bytes = hash_obj.digest()
  return base64.b64encode(digest_bytes).decode("utf-8")  # Decode to string


def _get_multipart_info(payload: bytes, header: bytes) -> tuple[bytes, bytes]:
  """Parse the multipart message and extract the object url and content.
  The combined message is in the format of:
  content-type: multipart/related; boundary=boundary_string
  --boundary_string
  
  content-type: application/json
  {
    "name": "object_url", "field_name": "field_value"
  }
  --boundary_string
  
  content-type: mime_type depending on the object type
  object_content
  --boundary_string--

  Args:
    payload: request body
    header:  request multipart/related header

  Returns:
    object_url: object url of the object to be uploaded
    object_content: object content of the object to be uploaded

  Raises:
    ValueError:
  """
  combined_msg = b"content-type: " + header + b"\n" + payload
  if len(combined_msg) > 100000:
    logger.debug(f"combined_msg is large: {len(combined_msg)}")
  else:
    logger.debug(f"combined_msg: {combined_msg}")
  msg_parser = parser.BytesParser()
  parsed_message = msg_parser.parsebytes(combined_msg)
  parts = list(parsed_message.walk())
  
  if len(parts) < 3:
    logger.error(f"Multipart message has {len(parts)} parts, expected at least 3.")
    raise ValueError(
        f"Multipart message has {len(parts)} parts, expected at least 3."
    )

  logger.debug(f"parts[1] content_type: {parts[1].get_content_type()}")
  logger.debug(f"parts[2] content_type: {parts[2].get_content_type()}")
  
  json_data = json.loads(parts[1].get_payload(decode=True))
  object_url = json_data.get("name").encode("utf-8")
  object_content = parts[2].get_payload(decode=True)
  
  return object_url, object_content


def encrypt_and_decrypt(
    mode: str, ciphertext: bytes, gcs_blob_path: bytes
) -> bytes:
  """Encrypts or decrypts data using Tink and Google Cloud KMS.

  Args:
      mode: The operation mode, either "encrypt" or "decrypt".
      gcp_project_id: The ID of the Google Cloud project.
      ciphertext: The data to encrypt or decrypt.
      gcs_blob_path: The path to the GCS blob.

  Returns:
      The encrypted or decrypted data.

  Raises:
      ValueError: If an unsupported mode is provided.
  """

  if mode == "encrypt":
    output_data = env_aead.encrypt(ciphertext, gcs_blob_path)
    logger.info("Data encrypted successfully.")
    return output_data
  elif mode == "decrypt":
    decrypted_content = env_aead.decrypt(ciphertext, gcs_blob_path)
    logger.info("Data decrypted successfully.")
    return decrypted_content
  else:
    raise ValueError(
        f'Unsupported mode {mode}. Please choose "encrypt" or "decrypt".'
    )


def _log_response(flow: http.HTTPFlow, heading: str = None):
  receiving = "<<<<"
  if heading:
    logger.debug(heading)
  logger.debug(f"---------- response begin-------------")
  logger.debug(f"{receiving}  flow id: {flow.id}")
  logger.debug(f"{receiving}  URL: {flow.request.pretty_url}")
  logger.debug(f"{receiving}  Method: {flow.request.method}")
  logger.debug(f"{receiving}  Headers: {flow.request.headers}")
  if flow.request.content:
    logger.debug(f"{receiving} content len: {len(flow.request.content)}")
  else:
    logger.debug(f"{receiving} content len: None")
  if flow.response:
    logger.debug(f"{receiving} response.status_code {flow.response.status_code}")
    logger.debug(f"{receiving} response.header {flow.response.headers}")
    if flow.response.content:
      logger.debug(f"{receiving} response.content len {len(flow.response.content)}")
    else:
      logger.debug(f"{receiving} response.content len: None")
  logger.debug(f"---------- response end-------------")


def _log_request(flow: http.HTTPFlow, heading: str = None):
  sending = ">>>>"
  if heading:
    logger.debug(heading)
  logger.debug(f"---------- request begin -------------")
  logger.debug(f"{sending}  flow id: {flow.id}")
  logger.debug(f"{sending}  URL: {flow.request.pretty_url}")
  logger.debug(f"{sending}  Method: {flow.request.method}")
  logger.debug(f"{sending}  Headers: {flow.request.headers}")
  if flow.request.content:
    logger.debug(f"{sending} content len: {len(flow.request.content)}")
  else:
    logger.debug(f"{sending} content len: None")
  logger.debug(f"---------- request end -------------")


def request(flow: http.HTTPFlow) -> None:
  """Intercepts and potentially modifies HTTP requests.

  This function specifically targets GCS upload requests, attempting to
  encrypt the content before it's uploaded.

  Args:
      flow: The HTTP flow object representing the request/response exchange.
  """
  try:
    _log_request(flow)
    if GCS_PROXY_DISABLE_ENCRYPTION:
      return
    if flow.request.pretty_url.startswith(
        "https://storage.googleapis.com/upload"
    ):
      logger.info("GCS Upload Request intercepted:")
      
      upload_type = flow.request.query.get("uploadType")
      object_name = flow.request.query.get("name")
      
      gcs_url = flow.request.pretty_url
      gcs_url = gcs_url.split("/v1/b/")[1]
      bucket_url = gcs_url.split("/o")[0]  # get the bucket and object url
      object_url = b""
      object_content = b""
      if upload_type == "multipart":
        multipart_header = (
            flow.request.headers.get("content-type")
            .replace("'", '"')
            .encode("utf-8")
        )
        object_url, object_content = _get_multipart_info(
            flow.request.content, multipart_header
        )
      elif upload_type == "media":
          object_url = object_name.encode("utf-8")
          object_content = flow.request.content
      else:
          logger.error(f"Unsupported upload type: {upload_type}")
          raise ValueError(f"Unsupported upload type: {upload_type}")
      
      gcs_path = (
          b"gs://"
          + bucket_url.encode("utf-8")
          + b"/"
          + object_url
      )
      logger.info(f"GCS Path of the upload request : {gcs_path}")

      # TODO @ericshen Temporay hack for gcloud storage cp.
      #      Set custom headers for original content length and original object md5 hash.
      #      We can set the md5 and content length to the original values in the response.
      flow.request.headers["gcs-proxy-original-md5-hash"] = _base64_md5hash(
          object_content
      )
      flow.request.headers["gcs-proxy-original-content-length"] = str(
          len(flow.request.content)
      )

      encrypted_object_content = encrypt_and_decrypt(
          "encrypt",
          object_content,
          gcs_path,
      )

      # Replace the object content with the encrypted object content
      # TODO @ericshen string replacement is probably ok for PoC.
      #      We should use a more robust way to do this.
      flow.request.content = flow.request.content.replace(
          object_content, encrypted_object_content
      )
      logger.info(f"Request content modified after encryption.")
  except Exception as e:
    logger.exception("Error processing request: %s", e)
    flow.response = http.Response.make(
        status_code=500,
        content=b"Encryption failed",
        headers={"Content-Type": "text/plain"},
    )


def response(flow: http.HTTPFlow) -> None:
  """Intercepts and potentially modifies HTTP responses.

  This function targets GCS download responses, attempting to decrypt the
  content that was previously encrypted.

  Args:
      flow: The HTTP flow object representing the request/response exchange.
  """
  # TODO @ericshen gcloud storage cat command calls /b to get the object metadata.
  #      It stores the size of the object in the response. It then calls /dowload 
  #      to get the object content, which is decrypted by the proxy and has a different size.
  #      The command downloads the decrypted object but gives an error for 
  #      mismaching size. We'd need to address this issue.
  #
  # if flow.request.method == "GET" and flow.request.pretty_url.startswith("https://storage.googleapis.com/storage/v1/b"):
  #     log_response(flow, "*** Response for /getObejct request")

  # TODO @ericshen hack for gcloud storage cp command
  #      Intercept the /upload response and change the md5 has to its original value
  try:
    _log_response(flow)
    if GCS_PROXY_DISABLE_ENCRYPTION:
      return     
      
    if flow.request.pretty_url.startswith(
        "https://storage.googleapis.com/upload"
    ):
      logger.info("GCS Upload response intercepted")
      original_md5_hash = flow.request.headers["gcs-proxy-original-md5-hash"]
      json_data = json.loads(flow.response.content)
      json_data["md5Hash"] = original_md5_hash
      flow.response.content = json.dumps(json_data).encode("utf-8")
      logger.debug(f"change response md5  {json_data['md5Hash']} to orginal md5 {original_md5_hash}.")
    if _is_download_response(flow):
      logger.info("GCS Download response intercepted:")
      gcs_url = flow.request.pretty_url
      gcs_url = gcs_url.split("/v1/b/")[1]
      bucket_url, object_url = (
          gcs_url.split("/o/")[0],
          gcs_url.split("/o/")[1],
      )  # get the bucket and object url
      # capture only object name and clean up the query string post ?
      object_url = object_url.split("?")[0]
      gcs_path = (
          b"gs://"
          + bucket_url.encode("utf-8")
          + b"/"
          + object_url.encode("utf-8")
      )
      encrypted_data = flow.response.content
      decrypted_content = encrypt_and_decrypt(
          "decrypt",
          encrypted_data,
          gcs_path,
      )
      # Update the response content with the decrypted content
      flow.response.content = decrypted_content

      # Update content lenght headers with new length of decrypted data
      flow.response.headers["X-Goog-Stored-Content-Length"] = str(
          len(flow.response.content)
      )

      content_length = len(flow.response.content)
      flow.response.headers["Content-Length"] = str(content_length)
      logger.debug(f"Update response content_length to the length of decrpyted content:  {content_length}")
      # gcloud storage cp command uses "range" in request
      range_value = f"bytes 0-{content_length-1}/{content_length}".encode("utf-8")

      if "range" in flow.request.headers:
        flow.request.headers["range"] = range_value

      if "Content-Range" in flow.response.headers:
        flow.response.headers["Content-Range"] = range_value

      # Update Google hash headers with decrypted contents md5 hash to pass the checksum validation.
      md5_hash = _base64_md5hash(decrypted_content)
      flow.response.headers["X-Goog-Hash"] = md5_hash
      logger.debug(f"Update response md5 hash to that of decrpyted content:  {md5_hash}") 
  except Exception as e:
    logger.exception("Error processing response: %s", e)
    flow.response = http.Response.make(
        status_code=500,
        content=b"Decryption failed",
        headers={"Content-Type": "text/plain"},
    )


def _is_download_response(flow: http.HTTPFlow) -> bool:
    if flow.request.pretty_url.startswith(
        "https://storage.googleapis.com/download"
    ):
        return True
    if re.search(f"storage/v1/b/.+/o", flow.request.pretty_url) and flow.request.method == "GET" and flow.response.status_code // 100 == 2:
        return True
    return False

# uncomment the following for python remote debugging
# debugpy.listen(("0.0.0.0", 5678))  # Listen on all interfaces, port 5678
# time.sleep(5) 
# debugpy.wait_for_client()  # Pause execution until debugger attaches

def start():
  """Entry point for starting the mitmproxy addon."""
  from mitmproxy import ctx
  ctx.master.addons.add(request)
  ctx.master.addons.add(response)
