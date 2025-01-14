#!/bin/bash

set -x

export OUTPUT_DIR="${OUTPUT_DIR:-gs://eshen-axlearn/fuji-$(date +%s)}"
export DATA_DIR="${DATA_DIR:-gs://axlearn-public/tensorflow_datasets}"
export CONFIG="${CONFIG:-fuji-7B-s1-b4}"
export SAVE_EVERY_N_STEPS="${SAVE_EVERY_N_STEPS:-100}"

if [[ -n "$https_proxy" ]]; then
  echo "GCS proxy enabled. Adding certificate to system store..."

  timeout=60  # 1 minute timeout
  start_time=$(date +%s)

  until [[ -f /proxy/certs/mitmproxy-ca-cert.pem ]] || (( $(date +%s) - start_time > timeout )); do
    echo "Waiting for /proxy/certs/mitmproxy-ca-cert.pem..."
    sleep 5
  done

  if [[ -f /proxy/certs/mitmproxy-ca-cert.pem ]]; then
    cp /proxy/certs/mitmproxy-ca-cert.pem /usr/local/share/ca-certificates/mitmproxy-ca-cert.crt
    update-ca-certificates -v
    ls -l /etc/ssl/certs | grep mitm
  else
    echo "Timeout waiting for /proxy/certs/mitmproxy-ca-cert.pem. Exiting."
    exit 1
  fi
fi

python3 -m axlearn.common.launch_trainer_main \
    --module=gke_fuji --config=$CONFIG \
    --trainer_dir=$OUTPUT_DIR --data_dir=$DATA_DIR --jax_backend=tpu