## Sample Go app(sending GCS requests) and GCS Proxy Sidecar Deployment

This example contains a sample Go application which makes calls to GCS api using GCS Go SDK and uses go-gcsproxy deployed as a sidecar in GKE(kubernetes).

### Key Features

* Sample Go application with 2 endpoints to send upload and download request to GCS.
* Open telemetry configurations in the sample GO application to monitor the performance.
* Kubernetes manifest file to deploy the go-gcsproxy as a sidecar container to the Sample Go application.

### Deployment Steps

* Deploy the sample Go application using the Dockerfile in the current directory.
* Deploy the Sample app along with the go-gcsproxy as a sidecar container in the kubernetes cluster by updating the `manifests/go-api.yaml` file.