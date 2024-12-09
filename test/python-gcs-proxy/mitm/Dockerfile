FROM python:3.9

WORKDIR /app

COPY requirement.txt /app

RUN pip install --no-cache-dir -r requirement.txt

# Copy the credentials directly in docker image. TODO: use k8s secret to mount the credentials.
COPY credentials.json /app
# Copy the mitmproxy code
COPY mitmproxy-gcs.py /app

CMD ["mitmdump", "--set", "confdir=/proxy/certs", "-s", "mitmproxy-gcs.py", "-w", "capture.mitm"]