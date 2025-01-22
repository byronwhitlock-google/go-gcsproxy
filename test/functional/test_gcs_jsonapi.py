# Copyright 2025 Google.
#
# This software is provided as-is, without warranty or representation for any use or purpose.
"""
Proxy funtioanl testing for GCS clients such as json API and SDK

Setup:
  Set the following the enviroment variables:
   -- PROXY_FUNC_TEST_BUCKET: GCS bucket for testing. Required
   -- https_proxy: Point to the proxy. Required
                   ie. https_proxy=https://localhost:8080
   -- CURL_CA_BUNDLE: Mitmproxy self-signed ca cert. Required

Usage:
  >>> pytest -v -s --log-cli-level=INFO test_gcs_clients.py

"""
import os
import pytest
import logging
import time
import json

import requests
import test_util
from urllib.parse import quote
from google.auth import default
from google.auth.transport.requests import Request


LOG_LEVEL_STR = os.environ.get("PROXY_FUNC_TEST_LOG_LEVEL", "INFO")
log_level = getattr(logging, LOG_LEVEL_STR.upper(), logging.INFO)
logging.basicConfig(level=logging.INFO,
                    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')

logger = logging.getLogger(__name__)

TEST_BUCKET = os.environ.get(
    "PROXY_FUNC_TEST_BUCKET",
    "gcs-proxy-func-test",
)
OBJECT_NAME = "func-test-object"
OBJECT_CONTENT = "testing object content"

TEST_UNIQUE_FOLDER = str(int(time.time() * 1000))
if os.environ.get("https_proxy"):
    TEST_UNIQUE_FOLDER += "-with-proxy"


GCS_TESTING_PATH = f"gs://{TEST_BUCKET}/{TEST_UNIQUE_FOLDER}"
logger.info(
    f"GCS testing path: {GCS_TESTING_PATH}  https_proxy: {os.environ.get('https_proxy')}")


@pytest.fixture(scope="module")
def setup_data():
    """Fixture to set up any necessary data or resources."""
    return {
        "original_object": OBJECT_CONTENT,
    }


def test_resumable_upload_byterange_download(setup_data):
    """Test case for resumable upload and chunked download."""

    test_id = test_resumable_upload_byterange_download.__name__
    object_url = test_util.generate_object_url(
        GCS_TESTING_PATH, OBJECT_NAME, test_id=test_id)
    expected = setup_data["original_object"]
    credentials, project_id = default(
        scopes=["https://www.googleapis.com/auth/cloud-platform"])
    credentials.refresh(Request())
    access_token = credentials.token
    object_name = object_url.replace(f"gs://{TEST_BUCKET}/", "")

    logger.info(f"Resumable upload {object_url}")
   
    # send POST request
    post_response = _resumable_upload_post(
        TEST_BUCKET, object_name, str(len(expected)), access_token)

    # send PUT request
    upload_id = post_response.headers["X-GUploader-UploadID"]
    content_range = f"bytes 0-{len(expected)-1}/{len(expected)}"
    put_response = _resumable_upload_put(
        TEST_BUCKET, object_name, content_range, upload_id, access_token, expected)
    
    logger.info(f"Download {object_url}")
    # get object metadata
    get_metadata_response = _get_object_metadata(
        TEST_BUCKET, object_name, access_token)
    content = json.loads(get_metadata_response.content.decode("utf-8"))
    metadata_uncrypted_content_length = content["metadata"]["x-unencrypted-content-length"]
    size = content["size"]
    generation = content["generation"]

    assert metadata_uncrypted_content_length == size and size == str(
        len(expected))

    # get object
    get_object_response = _get_object(
        TEST_BUCKET, object_name, generation, access_token, len(expected) - 1)

    actual = get_object_response.content.decode("utf-8")
    assert expected == actual


def _get_object(bucket, object_name, generation, access_token, end_index):
    encoded_object_name = object_name.replace("/", "%2F")
    url = f"https://storage.googleapis.com/download/storage/v1/b/{bucket}/o/{encoded_object_name}"
    headers = {
        "Accept": "*/*",
        "Authorization": f"Bearer {access_token}",
        "Range": f"bytes=0-{end_index}",
        "User-Agent": "TSL",
    }
    params = {
        "alt": "media",
        "generation": generation
    }

    response = requests.get(url, headers=headers, params=params)
    return response


def _get_object_metadata(bucket, object_name, access_token):
    encoded_object_name = object_name.replace("/", "%2F")
    url = f"https://storage.googleapis.com/storage/v1/b/{bucket}/o/{encoded_object_name}"
    headers = {
        "Accept": "application/json",
        "Authorization": f"Bearer {access_token}",
        "User-Agent": "TSL",

        "Accept-Encoding": "gzip, deflate",
        "Connection": "keep-alive",
        "Content-Length": "0",
        "X-Goog-Api-Client": "cred-type/u",
    }
    params = {
        "alt": "json",
        "projection": "noAcl"
    }
    response = requests.get(url, params=params, headers=headers)
    return response


def _resumable_upload_put(bucket, object_name, content_range, upload_id, access_token, data):
    url = f"https://www.googleapis.com/upload/storage/v1/b/{bucket}/o"
    params = {
        "uploadType": "resumable",
        "name": object_name,
        "upload_id": upload_id
    }
    headers = {
        "Accept": "*/*",
        "Authorization": f"Bearer {access_token}",
        "Content-Range": content_range,
        "Expect": "100-continue",
        "User-Agent": "TSL",
    }
    response = requests.put(url, headers=headers, params=params, data=data)
    return response


def _resumable_upload_post(bucket, object_name, content_length, access_token):
    url = f"https://www.googleapis.com/upload/storage/v1/b/{bucket}/o"
    params = {
        "uploadType": "resumable",
        "name": object_name
    }

    headers = {
        "Accept": "*/*",
        "Authorization": f"Bearer {access_token}",
        "Content-Length": "0",
        "Content-Type": "application/x-www-form-urlencoded",
        "Expect": "100-continue",
        "User-Agent": "TSL",
        "X-Upload-Content-Length": content_length
    }
    response = requests.post(url, params=params, headers=headers)
    return response


if __name__ == "__main__":
    pytest.main()
