# Copyright 2025 Google.
#
# This software is provided as-is, without warranty or representation for any use or purpose.
"""
Proxy funtioanl testing for gcloud cli

Setup:
  Set the following the enviroment variables:
   -- PROXY_FUNC_TEST_BUCKET: GCS bucket for testing. Required
   -- https_proxy: Point to the proxy. Required
                   ie. https_proxy=https://localhost:8080
   -- CURL_CA_BUNDLE: Mitmproxy self-signed ca cert. Required

Usage:
  >>> pytest -v -s --log-cli-level=INFO test_gcloud_cli.py


"""
import os
import pytest
import logging
import time
import subprocess

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

TEST_UNIQUE_FOLDER = str(int(time.time() * 1000)) + "-test-gcloud"
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
    expected = setup_data["original_object"]
    with open(source, 'w') as file:
        file.write(expected)
           

    logger.info(f"Copying {source} to {object_url}")
    result = subprocess.run(
        ["gcloud", "storage", "cp", source, object_url], capture_output=True, text=True)
    logger.info(f"Return Code: {result.returncode}")
    logger.info(f"Output: {result.stdout}")
    logger.info(f"Error: {result.stderr}", )
    assert result.returncode == 0

    result = subprocess.run(
        ["gcloud", "storage", "cat", object_url], capture_output=True, text=True)
    logger.info(f"Return Code: {result.returncode}")
    logger.info(f"Output: {result.stdout}")
    logger.info(f"Error: {result.stderr}", )

    assert result.returncode == 0
    assert expected == result.stdout.strip()

def test_curl_copy_large_file_command(setup_data):
    """Test case for curl command to upload """

    test_id = test_curl_copy_large_file_command.__name__
    testfile = "hugefile.bin"
    object_url = test_util.generate_object_url(
        GCS_TESTING_PATH, testfile, test_id=test_id)
    

    # Run the dd command using subprocess
    subprocess.run([
        "dd", 
        "if=/dev/zero",  # Input file (source of zeroes)
        f"of={testfile}",  # Output file (the file being created)
        "bs=1M",  # Block size of 1MB
        "count=100"  # Write 100 blocks (total 100MB)
    ],stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)

    
    result = subprocess.run(
        ["gcloud", "storage", "cp", testfile, object_url], capture_output=True, text=True)     

    logger.info(f"Copying {testfile} to {object_url}")
    logger.info(f"Return Code: {result.returncode}")
    logger.info(f"Output: {result.stdout}")
    logger.info(f"Error: {result.stderr}", )

    ## Check for successful upload ##
    assert result.returncode == 0



if __name__ == "__main__":
    pytest.main()
