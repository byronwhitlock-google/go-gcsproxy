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

### Performance Testing Summary

The following section shows the results from the performance testing conducted on the go-gcsproxy. The performance testing was executed using the [locust tool](https://locust.io/). The go-gcsproxy was load tested by deploying this proxy on a [GKE cluster](https://cloud.google.com/kubernetes-engine) as a [sidecar container](https://kubernetes.io/docs/concepts/workloads/pods/sidecar-containers/) to the simple flask application which calls the GCS API's and these GCS API calls are intercepted by the sidecar go-gcsproxy.

The following load testing profiles were used to perform the load testing:

| **Profiles** | **low** | **medium** | **high** |
|---|---|---|---|
| **GCS Proxy Container resources** | vcpu - 2<br>memory - 8<br>pods - 3 | vcpu - 32<br>memory - 128<br>pods - 3 | vcpu - 64<br>memory - 256<br>pods - 3 |
| **Max Concurrent User** | 10 | 250 | 500 |
| **File size** | 1MB | 100MB | N/A |

GKE Cluster Configurations include a main Flask application (with 2 vCPU and 8 GB memory resources) running alongside the proxy sidecar across all profiles. Client VM configurations include GCE VM instances with n2d-standard-80 configurations to run the Locust load testing tool.

* These graphs show how encryption, decryption, upload, and download times are affected by increasing file size and concurrency.
  - **Key takeaways:**
    - Upload/download times increase with file size: Bigger files take longer to move.
    - Encryption/decryption become less significant: While these also take time, they don't increase as dramatically as upload/download with larger files. This means encryption/decryption make up a smaller portion of the total time for big files.
    - Large files can cause OOM: 1GB files caused errors due to system limitations (running out of memory) on an 8 GB system, leading to failed requests. 500MB files worked fine with the same system resources.
    - A 10GB file was uploaded on a 32vcpu / 128 GB memory machine without failure.
![1-r](https://github.com/user-attachments/assets/3e8ed2cd-29a6-441b-8e3a-9c2cc435eb06)
![2-r](https://github.com/user-attachments/assets/201bd5dc-5e5c-4d43-a5c8-c9d194509d00)

* The following graph compares upload and download performance of a GCS object with and without a proxy. Tests used varying file sizes, consistent server resources (2 vCPUs and 8 GB memory), and two concurrency combinations. Results show minimal latency differences between proxy and no-proxy configurations for files under 500 MB. However, for larger files (>500 MB), the latency difference increases significantly.
![3-r](https://github.com/user-attachments/assets/6e42b4a3-41e0-4667-b825-d6bd629e9aee)

* **Proxy Overhead:** The proxy introduces latency, as expected. Upload latency with the proxy is generally higher than without it (e.g., row 3: 6.15 seconds vs. 2.32 seconds for 1MB file, 500 users). Download latency with the proxy is also consistently higher than without it.
* **Resource Scaling:** 
  - Increasing CPU and memory resources improves performance, both with and without the proxy. For instance, moving from 2vCPU/8GB to 32vCPU/128GB (rows 1-3 vs. 4-6) significantly boosts RPS and reduces latency, especially for smaller file transfers.
  - The performance with 32 vcpu and 64 vpcu are similar which means increasing resources beyond 32vcpu for proxy does not impact the performance. For example Row 4-6 with 32vcpu performance is similar to Row 7-9 with 64 vcpu.
  - A test with 16vCPUs/64GB shows very similar performance to a test with 32vCPUs/128GB. This suggests that after a certain resource capacity the proxy performance is not increasing by increasing the proxy capacity.
* **Concurrent User Impact:** As the number of concurrent users increases, latency increases substantially, particularly for larger file sizes. This effect is more pronounced with the proxy (e.g., row 12: 395 seconds with proxy vs. 384 seconds without for 100MB, 500 users).

The following table provides highlights of the results from different profiles of performance testing:
|   | **RPS (Proxy)** | **RPS (No Proxy)** | **Resources (vcpu and memory)** | **File Size (MB)** | **Max Concurrent Users (Users)** | **Encryption Time (seconds)** | **Avg Latency Upload with Proxy (seconds)** | **Avg Latency Upload without Proxy (seconds)** | **Decryption Time (seconds)** | **Avg Latency Download with Proxy (seconds)** | **Avg Latency Download without Proxy (seconds)** |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | 38 | 73 | 2 vcpu 8Gb | 1 | 10 | 0.04 | 0.27 | 0.18 | 0.05 | 0.21 | 0.08 |
| 2 | 75 | 203 | 2 vcpu 8Gb | 1 | 250 | 0.08 | 3.11 | 1.19 | 0.07 | 3.12 | 1.15 |
| 3 | 77.5 | 205 | 2 vcpu 8Gb | 1 | 500 | 0.08 | 6.15 | 2.32 | 0.07 | 6.19 | 2.29 |
| 4 | 44 | 62 | 32 vcpu 128Gb | 1 | 10 | 0.03 | 0.25 | 0.17 | 0.03 | 0.15 | 0.08 |
| 5 | 123 | 189 | 32 vcpu 128Gb | 1 | 250 | 0.02 | 1.62 | 1.1 | 0.02 | 1.56 | 1.02 |
| 6 | 117 | 193 | 32 vcpu 128Gb | 1 | 500 | 0.02 | 4.11 | 2.53 | 0.02 | 4.07 | 2.47 |
| 7 | 43 | 65 | 64 vcpu 256Gb | 1 | 10 | 0.03 | 0.25 | 0.17 | 0.03 | 0.15 | 0.08 |
| 8 | 122 | 189 | 64 vcpu 256Gb | 1 | 250 | 0.02 | 2 | 1.3 | 0.02 | 1.9 | 1.2 |
| 9 | 123 | 188 | 64 vcpu 256Gb | 1 | 500 | 0.02 | 3.9 | 2.6 | 0.02 | 3.8 | 2.5 |
| 10 | 1.2 | 1.9 | 2 vcpu 8Gb | 100 | 10 | 0.5 | 11.4 | 9.7 | 0.1 | 3.6 | 4.2 |
| 11 | 2.5 | 0.8 | 2 vcpu 8Gb | 100 | 250 | 0.3 | 159 | 161 | 0.1 | 168 | 185 |
| 12 | 0.4 | 1.5 | 2 vcpu 8Gb | 100 | 500 | 0.6 | 337 | 325 | 0.1 | 395 | 384 |
| 13 | 1.5 | 1.4 | 32 vcpu 128Gb | 100 | 10 | 0.3 | 10.46 | 12.94 | 0.1 | 3.52 | 5.15 |
| 14 | 2.2 | 0.8 | 32 vcpu 128Gb | 100 | 250 | 0.2 | 159.8 | 161.2 | 0.1 | 169 | 184 |
| 15 | 1.6 | 0.7 | 32 vcpu 128Gb | 100 | 500 | 0.2 | 321.72 | 313.16 | 0.1 | 368.5 | 529.12 |
| 16 | 1.7 | NA | 64 vcpu 256Gb | 100 | 250 | 0.2 | 156 | NA | 0.1 | 168 | NA |
| 17 | 0.4 | 0.8 | 2 vcpu 8Gb | 250 | 10 | 0.8 | 31.2 | 30.3 | 0.2 | 8 | 9.5 |
| 18 | 0.2 | 0.5 | 2 vcpu 8Gb | 500 | 5 | 1.8 | 39.4 | 36.9 | 0.3 | 10 | 10.5 |
| 19 | 0 | 0.1 | 2 vcpu 8Gb | 1000 | 5 | 2.3 | 168.6 | 85.7 | 0.7 | 46 | 31.3 |

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
