# AX learn tests

To run these test, you can use the included DOCKERFILE via docker or podman.

## Certificates
 - Start the the go-gcsproxy 
 - Use an environment variable for the generated certs so we can pick them up for the test client.
 - gcs-goproxy with env var set: PROXY_CERT_PATH=&lt;root path&gt;/go-gcsproxy/test
  
## GCP authentication
 - The container expects the ADC key in ~/.config/gcloud
 - This is populated automatically if you run `sudo gcloud application-default login` on the host.

## Run tests
- Login to Gcloud
  - `sudo gcloud application-default login`

- Build the container locally
  - `sudo podman build . -t go-gcsproxy`
- Make sure you have access to a bucket. Replace &lt;bucketname&gt; with the name of that bucket.
  - `sudo podman run -e PROXY_FUNC_TEST_BUCKET=<bucketname> --mount type=bind,source=${HOME}/.config/gcloud,target=/app/.config/gcloud localhost/go-gcsproxy`

## GO-GCSPROXY
For now, GO-GCSPROXY needs to use a patched go, net/http is updated to handle TE:identity. 

Follow the instruction below to build a patched go command and toolchain:
1. ```git clone git@github.com:golang/go.git```
2. ```git checkout go1.23.0``` go-gcsproxy uses 1.23
3. Make changes to [your-go-repo-root]/src/net/http/transfer.go. Search "eshen" in [here](./go-net-http-patch/transfer.go) as refernce. 
4. Run ```make.bash``` under [your-go-repo-root]/src/. It generates go command under bin/ and toolchain under pkg/
5. Add [your-go-repo-root]/bin/ into PATH env, so the patched go command and toolchain will be used when you launch go-gcsproxy.
6. Go to go-gcsproxy directory and launch the proxy(make uses the patched go). 
