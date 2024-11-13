# go-gcsproxy

## Janiculum
Encrypting Reverse proxy for Google Cloud Storage.

## Description
### [P0] Requirement 1 (MVP)
- [x] Small binary written in a compiled language (Golang preferred)
- [ ] Use GCP KMS for keys
- [ ] Use Tink for encryption, following existing guidance
- [ ] Follow BIGFOOT internal guidance for encryption
- [ ] BIGFOOT intranet page, will only load if on BIGFOOT VPN
- [ ] Only use approved algorithms, such as SHA, AES, RSA, ECDSA.
- [ ] Use appropriate bit sizes, such as SHA-256, RSA-2048, ECDSA-256.
- [x] Must be scalable
- [x] Must work behind a load balancer
- [x] Must work as a sidecar deployment
- [ ] ~~Work with arbitrary GCS calls~~ *NOT POSSIBLE*
- [ ] Desired utilities to test: gcloud, gsutil, tensorflow, python SDK, go SDK, cURL
- [ ] Work for both HTTP and HTTPS
- [ ] Provide Terraform deployment automation template
- [x] Proxy should check if the traffic is bound for GCS and just pass along all other traffic
- [ ] Test using OS configuration environment variable NO_PROXY to verify non GCS traffic can be directed to NOT use the proxy. Most SDKs also allow similar configuration.
- [ ] Safe logging
- [ ] No keys or data can be leaked in logging, including to cloud logging
- [ ] Configurable logging: debug, error, warning, info, etc.
