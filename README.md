# go-gcsproxy

## Janiculum

Encrypting Reverse proxy for Google Cloud Storage.

## Description

This project provides an encrypting reverse proxy for Google Cloud Storage
(GCS). It is designed to intercept and modify HTTP/HTTPS traffic for GCS
operations, encrypting data before upload and decrypting it after download.
This adds an additional layer of security to the existing GCS service-side
encryption offerings. This is especially useful for organizations with strict
security and privacy requirements, such as those who want to prevent
even Google from having access to their data.

### Features

  * Transparent Encryption/Decryption: Ensures that data uploaded to GCS is
    automatically encrypted using Google Cloud KMS and Tink, while downloaded
    data is seamlessly decrypted.
  * Man-in-the-Middle (MITM) Proxy: Employs MITM proxy to intercept and modify
    HTTP/HTTPS traffic for GCS operations.
  * Tink Library: Leverages the Tink library for robust cryptographic operations
    and secure key management.
  * Easy-to-Use: Works out of the box with `gsutil` and `gcloud` commands,
    requiring no complex configurations.
  * Key Management:
      * Uses GCP KMS for key management.
      * Allows for specifying an encryption key per bucket using key-value pairs.
  * Compliance:
      * Employs only approved algorithms (SHA, AES, RSA, ECDSA) with appropriate
        bit sizes (SHA-256, RSA-2048, ECDSA-256).
  * Scalability: Designed to be scalable and work behind a load balancer.
  * Deployment: Can be deployed as a sidecar deployment.
  * Logging:
      * Safe logging practices prevent leaks of keys or data.
      * Configurable logging levels (debug, error, warning, info, etc.).

### Usage (Server)

1.  Build the `go-gcsproxy` binary:
    ```bash
    make
    ```
2.  Configure the proxy's behavior through environment variables. For a
    comprehensive list of available options, refer to the `Makefile`.
3.  Run the proxy:
    ```bash
    ./go-gcsproxy -debug=1  \
      -kms_resource_name=projects/YOUR_PROJECT_ID/locations/global/keyRings/YOUR_KEYRING/cryptoKeys/YOUR_CRYPTO_KEY
      -cert_path=/your/path/to/certs # mitmproxy-ca.pem is automatically generated on first run of proxy
    ```
4. (optional) configure environment variables for `GCP_KMS_RESOURCE_NAME, PROXY_CERT_PATH, SSL_INSECURE, DEBUG_LEVEL, GCP_KMS_BUCKET_KEY_MAPPING`

### Usage (Client)    
To use `gsutil` or `gcloud` with the `go-gcsproxy`, you need to configure them to
use the proxy and trust the proxy's CA certificate.

### Setting Proxy Environment Variables

Set the following environment variables to direct  `gcloud` traffic
through the proxy:

```bash
export https_proxy=http://127.0.0.1:9080
export http_proxy=http://127.0.0.1:9080
export HTTPS_PROXY=http://127.0.0.1:9080
export REQUESTS_CA_BUNDLE=/your/path/to/certs/mitmproxy-ca.pem
gcloud config set custom_ca_certs_file $REQUESTS_CA_BUNDLE
```
For detailed testing instructions and information on setting up a test
environment, please refer to the [testing documentation](./test/README.md).

### Encryption Key per GCS Path
By default, every request to GCS will be encryted including requests to public datasets.

This optional feature allows for more granular control over encryption keys by enabling
the specification of different KMS keys for different GCS paths (buckets or
sub-paths within buckets). 

The `GCP_KMS_BUCKET_KEY_MAPPING` parameter (or `-gcp_kms_bucket_key_mappings` command-line flag) accepts a key-value encoded string to map GCS paths to KMS keys.

**Buckets not listed in  `GCP_KMS_BUCKET_KEY_MAPPING` will pass-thru to GCS unencrypted**

**Example:**

GCP_KMS_BUCKET_KEY_MAPPING="bucket1:projects/project1/locations/global/keyRings/keyring1/cryptoKeys/key1,bucket2/path/to/data:projects/project2/locations/global/keyRings/keyring2/cryptoKeys/key2"

This example maps `bucket1` to `key1` and `bucket2/path/to/data` to `key2`.

## Roadmap

  * P0 (MVP): 
      * Meet core requirements including basic encryption/decryption, key
        management, and compatibility with common tools like `gsutil` and
        `gcloud`.
      * Internal Google testing.
  * P1:
      * Enhanced deployment options (sidecar in GKE, controller-managed
        annotation).
      * Support for JSON and gRPC APIs.
      * Dynamic configuration updates.
  * P2:
      * Integration with GCSFUSE.
      * Potential performance optimizations for TPUs.
      * Terraform deployment templates for non-GKE deployments.

## Considerations

  * Dependencies:
      * Tink library
  * Risks and Mitigations:
      * Potential slow adoption due to organizational factors.
  * Support and Tools:
      * Best-effort support.

## Limitations

* Uploads are currently limited to 100MB in `gcloud`.
* Streaming uploads are not fully supported.
* Resumable uploads are not fully supported.

These limitations will be addressed by the upcoming feature request for streaming uploads.

## New Feature Request: Streaming Uploads

### Algorithm for Streaming Uploads

This section outlines the proposed algorithm for supporting streaming uploads to
GCS via the `go-gcsproxy`, enhancing its functionality and compatibility with
various data transfer scenarios.

#### Upload Process

GCS handles streaming uploads through a series of distinct requests:

1.  **Initiate Upload:** A POST request to the bucket path initiates the upload
    and returns a unique upload ID.
2.  **Upload Chunks:** Subsequent PUT requests send individual chunks of data.
    Each request includes the upload ID, the byte range of the chunk being
    uploaded, and the chunk data itself. GCS appends each chunk to the
    composite object in the order received.
3.  **Finalize Upload:** (Implicit) Once all chunks are uploaded, GCS
    automatically finalizes the composite object.

Currently, the proxy handles the initial POST and the first PUT request. This
enhancement focuses on enabling the proxy to handle the subsequent PUT requests
for the remaining chunks, providing comprehensive support for streaming uploads.

#### Encryption and Metadata

  * Each chunk, corresponding to a single PUT request from the client, is
    encrypted independently by the proxy before upload.
  * The encrypted length of each chunk is stored as custom metadata in the final
    composite object to facilitate accurate decryption during download.
  * Metadata format:
    ```
    x-chunk-len-1: <length of chunk 1>
    x-chunk-len-2: <length of chunk 2>
    x-chunk-len-3: <length of chunk 3>
    ...
    x-unencrypted-length: ...
    x-md5-hash: ...
    ```

#### Metadata Management

  * **Persistent Cache Update:** The proxy's local cache, storing bucket
    information by upload ID, is extended to store encrypted offsets of each
    chunk associated with that ID, enabling the proxy to track the composite
    object's structure.
  * **Metadata Update Trigger:** The crucial metadata update, containing the
    `x-chunk-len-` entries, occurs only after the final chunk is processed and
    the composite object exists. The proxy is modified to detect the final
    chunk upload response from GCS, triggering the metadata update.

#### Download Process

1.  **Download and Buffer:** The entire object is downloaded and buffered.
2.  **Chunk Identification:** Custom metadata is parsed to determine the
    starting offset and length of each encrypted chunk.
3.  **Decryption and Concatenation:** Each chunk is decrypted independently, and
    the decrypted chunks are concatenated to reconstruct the original file.
4.  **Data Return:** The final, decrypted object is returned to the client.