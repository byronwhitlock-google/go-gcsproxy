# go-gcsproxy

## Janiculum
Encrypting Reverse proxy for Google Cloud Storage.

## Description
### [P0] Requirement 1 (MVP)
- [x] Small binary written in a compiled language (Golang preferred)
- [x] Use GCP KMS for keys
- [x] Use Tink for encryption, following existing guidance
- [ ] Follow BIGFOOT internal guidance for encryption
- [ ] BIGFOOT intranet page, will only load if on BIGFOOT VPN
- [x] Only use approved algorithms, such as SHA, AES, RSA, ECDSA.
- [x] Use appropriate bit sizes, such as SHA-256, RSA-2048, ECDSA-256.
- [x] Must be scalable
- [x] Must work behind a load balancer
- [x] Must work as a sidecar deployment
- [ ] ~~Work with arbitrary GCS calls~~ *NOT POSSIBLE*
- [ ] Desired utilities to test:
    - [x] `gcloud`
    - [x] `gsutil`
    - [ ] `tensorflow`
    - [x] `python SDK`
    - [ ] `go SDK` 
    - [ ] `cURL`
  - [x] Support Standard Multi-part Upload 
  - [ ] ~~Support XML API~~
  - [x] Support JSON API
- [x] Work for both HTTP and HTTPS
- [ ] Provide Terraform deployment automation template
- [x] Proxy should check if the traffic is bound for GCS and just pass along all other traffic
- [ ] Test using OS configuration environment variable NO_PROXY to verify non GCS traffic can be directed to NOT use the proxy. Most SDKs also allow similar configuration.
- [x] Safe logging
- [x] No keys or data can be leaked in logging, including to cloud logging
- [x] Configurable logging: debug, error, warning, info, etc.

### Docker command to run go proxy

```sh
sudo docker run -e GCP_KMS_RESOURCE_NAME=projects/axlearn/locations/global/keyRings/proxy/cryptoKeys/proxy-kek -it --rm us-docker.pkg.dev/axlearn/gcs-proxy/go-mitmproxy:v1
```
