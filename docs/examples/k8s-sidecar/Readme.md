## Sample Go app(sending GCS requests) and GCS Proxy Sidecar Deployment

This example contains a sample Go application which makes calls to GCS api using GCS Go SDK and uses go-gcsproxy deployed as a sidecar in GKE(kubernetes).

### Key Features

* Sample Go application with 2 endpoints to send upload and download request to GCS.
* Open telemetry configurations in the sample GO application to monitor the performance.
* Kubernetes manifest file to deploy the go-gcsproxy as a sidecar container to the Sample Go application.

### Deployment Steps

* Deploy the sample Go application using the Dockerfile in the current directory.
* Deploy the Sample app along with the go-gcsproxy as a sidecar container in the kubernetes cluster by updating the `manifests/go-api.yaml` file.

### Open Telemetry Encryption Time captured in GCP Cloud Monitoring
The following graph shows the encryption time captured in cloud monitoring. **Note:** The metrics are only visible once the Open telemetry collector is setup on the GKE cluster and the sample go app along with go-gcsproxy is deployed.
![8qLy5dYwG7R6Kay](https://github.com/user-attachments/assets/d4042345-303e-42f6-9de5-8acceee0c1f7)

### Open Telemetry Decryption Time captured in GCP Cloud Monitoring
The following graph shows the decryption time captured in cloud monitoring. **Note:** The metrics are only visible once the Open telemetry collector is setup on the GKE cluster and the sample go app along with go-gcsproxy is deployed.
![8wNtqyKkmrXfscX](https://github.com/user-attachments/assets/ef0e7333-6223-44f9-9148-5634fed83885)
