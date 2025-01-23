FROM python:3.11 as base

WORKDIR /app

# install gcloud for testing
RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg  add - && apt-get update -y && apt-get install google-cloud-cli -y

COPY requirement.txt .
RUN pip install --no-cache-dir -r requirement.txt

#update the keystore
# start gcs-goproxy with env var set: PROXY_CERT_PATH=<root path>/go-gcsproxy/test
# this will generate the pem below.
ADD mitmproxy-ca.pem /usr/local/share/ca-certificates/go-gcsproxy-ca.crt
RUN chmod 644 /usr/local/share/ca-certificates/go-gcsproxy-ca.crt && update-ca-certificates

FROM base as runner
COPY . . 
ENV SSL_CERT_FILE=/usr/local/share/ca-certificates/go-gcsproxy-ca.crt
ENV CURL_CA_BUNDLE=/usr/local/share/ca-certificates/go-gcsproxy-ca.crt
ENV REQUESTS_CA_BUNDLE=/usr/local/share/ca-certificates/go-gcsproxy-ca.crt

ENV https_proxy=http://host.docker.internal:9080
ENV GOOGLE_APPLICATION_CREDENTIALS=/root/.config/gcloud/application_default_credentials.json

CMD [ "pytest", "-v", "-s", "--log-cli-level=INFO"]