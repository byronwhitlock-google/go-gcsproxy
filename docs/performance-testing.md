# Performance Testing

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
