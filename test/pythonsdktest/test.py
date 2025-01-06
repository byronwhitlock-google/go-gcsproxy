import os

import google.cloud.storage as storage

# Configure Google Cloud Storage
storage_client = storage.Client()

# Uncomment the following to test locally
os.environ["https_proxy"] = "http://127.0.0.1:9080"
os.environ["REQUESTS_CA_BUNDLE"] = "/Users/lkolluru/working-dir/apple/go-gcsproxy/test/mitmproxy-ca-cert.pem"

bucket_name = "ehorning-axlearn"
blob_name = "10mb.txt"
destination_blob_name = "10mb-go-res.txt"
bucket = storage_client.bucket(bucket_name)
blob = bucket.blob(destination_blob_name)
blob.upload_from_filename(blob_name)

print(f'File {blob_name} uploaded to {bucket_name}/{destination_blob_name}')

destination_file_name = "10mbres.txt"
blob.download_to_filename(destination_file_name)