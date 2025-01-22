# Functional testing

Functional testing is intented to test various clients with the proxy. The targed clients include:
* [Axlearn](./test_axlearn_tf.py) -- A ML framework that uses TF libaries to access data in GCS.
* [JSON API](./test_gcs_jsonapi.py) -- Test GCS JSON API directly.
* GCS SDK(TBD) -- Test with GCS python client SDK.

To run these test, you can use the included DOCKERFILE via docker or podman.

## Certificates
 - Start the the [go-gcsproxy](../../README.md#usage-server) 
 - Use an environment variable for the generated certs so we can pick them up for the test client.
 - gcs-goproxy with env var set: PROXY_CERT_PATH=&lt;root path&gt;/go-gcsproxy/test
  
## GCP authentication
 - The container expects the ADC key in ~/.config/gcloud
 - This is populated automatically if you run `sudo gcloud application-default login` on the host.

## Run tests
- Login to Gcloud
  - `sudo gcloud application-default login`

- Build the container locally
  - `sudo podman build . -t proxy-test-client`
- Make sure you have access to a bucket. Replace &lt;bucketname&gt; with the name of that bucket.
  - `sudo podman run -e PROXY_FUNC_TEST_BUCKET=<bucketname> --mount type=bind,source=${HOME}/.config/gcloud,target=/app/.config/gcloud localhost/proxy-test-client`  
