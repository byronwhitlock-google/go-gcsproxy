# AX learn tests

To run these test, you can use the included DOCKERFILE via docker or podman.

## Certificates
 - Start the the go-gcsproxy 
 - Use an environment variable for the generated certs so we can pick them up for the test client.
 - gcs-goproxy with env var set: PROXY_CERT_PATH=<root path>/go-gcsproxy/test
  
## GCP authentication
 - The container expects the ADC key in ~/.config/gcloud
 - This is populated automatically if you run `sudo gcloud application-default login` on the host.

## Run tests
- Login to Gcloud
  - `sudo gcloud application-default login`

- Build the container locally
  - `sudo podman build . -t go-gcsproxy`
- Make sure you have access to a bucket. Replace <bucketname> with the name of that bucket.
  - `sudo podman run -e PROXY_FUNC_TEST_BUCKET=<bucketname> --mount type=bind,source=${HOME}/.config/gcloud,target=/app/.config/gcloud localhost/go-gcsproxy`
