
  

  

# Overview

  

The purpose of the acceptance testing is to ensure that Axlearn(ML framework) workloads work with the gcs proxy. An Axlearn [fuji-7B training job](https://github.com/apple/axlearn/blob/main/axlearn/experiments/text/gpt/fuji.py) (on c4 dataset) is run on a single v4 pod slice with the gcs proxy as a sidecar container in a GKE cluster. The encrypt is performed by the proxy on checkpoints(every 100 steps) and the decrypt on resuming the training job from restoring the most recent checkpoint.

  

  

# Instructions

  

To run the fuji-7B training job with the proxy as a sidecar, following the steps below:

  

  

1. Create a standard zone(us-central2-b) GKE cluster and a nodepool with TPU(ct4p-hightpu-4t 2x2x4). Example:

```
gcloud container node-pools create axlearn-fuji \
--location=us-central2-b \
--cluster=eshen-gcs-proxy \ # change to your cluster
--node-locations=us-central2-b \
--machine-type=ct4p-hightpu-4t \
--tpu-topology 2x2x4 \
--spot \
--num-nodes=4 \
--scopes=https://www.googleapis.com/auth/cloud-platform
```

  

2. Build docker images for both Proxy and Axlearn following [docker](#build-docker-images) section. 
3. Connect to your cluster
```
gcloud container clusters get-credentials eshen-gcs-proxy \ # change to your cluster
--zone us-central2-b \
--project cool-machine-learning # change to your project
```
3. Install the latest JobSet in the cluster
```
kubectl apply --server-side -f https://github.com/kubernetes-sigs/jobset/releases/download/v0.7.2/manifests.yaml
``` 
3. Run the training job by applying [multi-host-job-gcs-proxy.yaml](./multi-host-job-gcs-proxy.yaml)

  

```
kubctl apply -f multi-host-job-gcs-proxy.yaml
```
4. Once the job starts running, you'd see the log like [this](./run-restore-with-proxy.log). Wait until several checkpoints have been written to your output GCS bucket. Then stop and remove the training job.

```
kubctl delete jobset fuji-multihost-job-gcs-proxy
```
5. Run step 3 to restart the job to restore from the most recent checkpoint.

  
# Build Docker Images
## Proxy
```
$> cd <repo-root>
$> docker build --platform=linux/amd64 -f ./Dockerfile.go123patch  -t us-docker.pkg.dev/cool-machine-learning/eshen-gcs-proxy/go-gcs-proxy-patch . # change to your own target
$> docker push us-docker.pkg.dev/cool-machine-learning/eshen-gcs-proxy/go-gcs-proxy-patch
```

## Axlearn
```
$> cd <repo-root>/test/acceptance
$> docker build --platform=linux/amd64 -t us-docker.pkg.dev/cool-machine-learning/eshen-gcs-proxy/axlearn/fuji .  # change to your own target
$> docker push us-docker.pkg.dev/cool-machine-learning/eshen-gcs-proxy/axlearn/fuji
```


# Configuration

Make sure you use the images that you have built from the [docker](#build-docker-images) section in the [manifest](./multi-host-job-gcs-proxy.yaml). There are also environment variables you can control as below

  

## Proxy Container(sidecar)

```
# - name: GCS_PROXY_DISABLE_ENCRYPTION # Disable encryption by settting to true
# value: "true"
- name: DEBUG_LEVEL # Default is 0 which is INFO
  value: "1"
- name: PROXY_CERT_PATH
  value: "/proxy/certs" # don't change this
- name: GCP_KMS_BUCKET_KEY_MAPPING # change it to your kms key
  value: "eshen-gcs-proxy-acceptance:projects/cool-machine-learning/locations/global/keyRings/proxy/cryptoKeys/proxy-kek"
```

## Axlearn Container
```
- name: CONFIG
  value: fuji-7B-s1-b32 # don't change
- name: OUTPUT_DIR
  value: "gs://eshen-gcs-proxy-acceptance" # change it to your GCS bucket. It should be the bucket in the KEP_MAPPING env for the proxy
- name: https_proxy # don't chnage

  value: "http://127.0.0.1:9080"
- name: GCS_RETRY_CONFIG_MAX_RETRIES # don't change
  value: "0"
```

  
  
  

# Findings

We've run the testing with and without encryption(via GCS_PROXY_DISABLE_ENCRYPTION env). The impact of the encrypt/decrypt on checkpointing seems to be reasonable.

 
||Write Checkpoint(avg)|Restore Checkpoint(avg)|Logs
|--|--|--|--
|Encryption|~60s|~40s|[run-restore-with-proxy.log](./run-restore-with-proxy.log)
|No Encryption |~50s|~33s|[run-restore-without-proxy.log](./run-restore-without-proxy.log)
 
Notes:

  

1. To get the time that it takes to write a checkpoint, look for the following line in the log

```
23:03:15.159983 135894283523776 checkpointer.py:510] Serialization of gs://eshen-gcs-proxy-acceptance/checkpoints/step_00000100 completed in 57.10629372299809 seconds.
```

2. To get the time that it takes to restore a checkpoint, look for the following lines in the log and calculate the time differences between the timestamp

```
**23:10:50.390969** 132033253665664 checkpointer.py:539] Restoring checkpoint from directory gs://eshen-gcs-proxy-acceptance/checkpoints/step_00000200

**23:11:29.364421** 132033253665664 checkpointer.py:1088] Restored state from ckpt at step 200
```
3. A sample checkpoint step has about 288 objects (65.22GiB), which is typical. 