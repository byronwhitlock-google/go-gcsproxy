# Copyright 2025 Google.
#
# This software is provided as-is, without warranty or representation for any use or purpose.
"""
Proxy funtioanl testing for clients using SDK

Setup:
  Set the following the enviroment variables:
   -- PROXY_FUNC_TEST_BUCKET: GCS bucket for testing. Required
   -- https_proxy: Point to the proxy. Required
                   ie. https_proxy=https://localhost:8080
   -- CURL_CA_BUNDLE: Mitmproxy self-signed ca cert. Required

Usage:
  >>> pytest -v -s --log-cli-level=INFO test_sdk.py


"""
import os
import pytest
import logging
import time
from google.cloud import storage

import test_util


LOG_LEVEL_STR = os.environ.get("PROXY_FUNC_TEST_LOG_LEVEL", "INFO")
log_level = getattr(logging, LOG_LEVEL_STR.upper(), logging.INFO)
logging.basicConfig(level=logging.INFO,
                    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')

logger = logging.getLogger(__name__)

TEST_BUCKET = os.environ.get(
    "PROXY_FUNC_TEST_BUCKET",
    "gcs-proxy-func-test",
)
OBJECT_NAME = "func-test-object"
OBJECT_CONTENT = "testing object content"

TEST_UNIQUE_FOLDER = str(int(time.time() * 1000)) + "-test-sdk"
if os.environ.get("https_proxy"):
    TEST_UNIQUE_FOLDER += "-with-proxy"


GCS_TESTING_PATH = f"gs://{TEST_BUCKET}/{TEST_UNIQUE_FOLDER}"
logger.info(
    f"GCS testing path: {GCS_TESTING_PATH}  https_proxy: {os.environ.get('https_proxy')}")


@pytest.fixture(scope="module")
def setup_data():
    """Fixture to set up any necessary data or resources."""
    return {
        "original_object": OBJECT_CONTENT,
    }


def test_cli_copy_cat(setup_data):
    """Test case for gcloud storage cp to upload and cat to download """

    test_id = test_cli_copy_cat.__name__
    source = "/tmp/source"
    object_url = test_util.generate_object_url(
        GCS_TESTING_PATH, OBJECT_NAME, test_id=test_id)
    object_path = object_url.replace(f"gs://{TEST_BUCKET}/", "")
    expected = setup_data["original_object"]
    with open(source, 'w') as file:
        file.write(expected)

    client = storage.Client()
    bucket = client.bucket(TEST_BUCKET)
    blob = bucket.blob(object_path)
    logger.info(f"Copying {source} to {object_url}")
    blob.upload_from_filename(source)

    logger.info(f"Downloading {object_url}")
    actual = blob.download_as_bytes().decode("utf-8")
    
    assert expected == actual





if __name__ == "__main__":
    pytest.main()
