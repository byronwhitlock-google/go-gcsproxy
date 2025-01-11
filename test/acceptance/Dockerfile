FROM python:3.10
WORKDIR /workspace
RUN pip install "jax[tpu]==0.4.34" -f https://storage.googleapis.com/jax-releases/libtpu_releases.html
RUN pip install "axlearn[core,gcp] @git+https://github.com/apple/axlearn.git@1af2ba8714a0a88aaf839841291b12f87e132eb4"
COPY run-fuji.sh .
COPY gke_fuji.py .
CMD bash run-fuji.sh