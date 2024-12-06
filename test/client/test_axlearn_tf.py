"""
Proxy funtioanl testing for Axlearn(tensforflow) libraries to interface GCS

Setup:
  Set the following the enviroment variables
   -- PROXY_FUNC_TEST_BUCKET: GCS bucket for testing. Required
   -- https_proxy: Point to the proxy. Required
                   ie. https_proxy=https://localhost:8080
 
   
For tensorflow(tf.io, tf.data), you'd need to add the mitmproxy self-signed ca cert to the system store. 
   -- Linux: use update-ca-certficates
         (https://manpages.ubuntu.com/manpages/xenial/man8/update-ca-certificates.8.html)
   -- Mac: Keychain Access -> Certificates

Usage:
  >>> pytest -v -s --log-cli-level=INFO test_axlearn_tf.py

"""
import os
import pytest
import logging
import time
import test_util
import tensorflow as tf
from typing import Sequence
from axlearn.common import file_system as fs
# from axlearn.common import input_tf_data as tf_data
from axlearn.common.config import config_for_function
from axlearn.common.input_tf_data import (tfds_dataset, tfrecord_dataset)
import uuid
import tensorstore as ts
import numpy as np

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

TEST_UNIQUE_FOLDER = str(int(time.time() * 1000))
GCS_TESTING_PATH = f"gs://{TEST_BUCKET}/{TEST_UNIQUE_FOLDER}"
logger.info(f"GCS testing path: {GCS_TESTING_PATH}")


@pytest.fixture(scope="module")
def setup_data():
    """Fixture to set up any necessary data or resources."""
    return {
        "original_object": OBJECT_CONTENT,
    }


@pytest.mark.skip(reason="temp")
def test_axlearn_fileio_copy(setup_data):
    """Test case for axlearn file_io.copy()"""
    from axlearn.common import file_system as fs
    source = "/tmp/source"
    object_url = test_util.generate_object_url(GCS_TESTING_PATH, OBJECT_NAME)

    expected = setup_data["original_object"]
    with open(source, 'w') as file:
        file.write(expected)

    logger.info(f"Copying {source} to {object_url}")
    fs.copy(source, object_url, overwrite=True)

    logger.info(f"Getting {object_url}")
    with fs.open(object_url) as f:
        actual = f.read()
    assert expected == actual

# @pytest.mark.skip(reason="temp")


def test_tf_io_gfile_write_read(setup_data):
    """Test case for tf.io.gfile.GFile.write()"""
    expected = setup_data["original_object"]
    object_url = test_util.generate_object_url(GCS_TESTING_PATH, OBJECT_NAME)

    logger.info(f"Creating {object_url}")
    with tf.io.gfile.GFile(object_url, "w") as f:
        f.write(expected)

    logger.info(f"Getting {object_url}")
    with tf.io.gfile.GFile(object_url, "r") as f:
        actual = f.read()
    assert expected == actual

# @pytest.mark.skip(reason="temp")


def test_tf_data_write_read(setup_data):
    """Test case for tf.data.TFRecordDataset which is used by axlearn input_tf_data.tfrecrod_dataset"""
    test_id = uuid.uuid4()

    expected = {
        "texts": ["a", "b", "c", "d"]
    }

    # Create Dataset
    ds_fn = test_util.build_ds_fn(texts=expected["texts"])
    dataset = ds_fn()

    # Serialize Dataset
    serialized_dataset = dataset.map(test_util.tf_serialize_example)

    # Write to GCS in TFRecord Format (Sharded into two files)
    gcs_path_1 = test_util.generate_object_url(
        GCS_TESTING_PATH, OBJECT_NAME + "-shard1.tfrecord", test_id=test_id)
    gcs_path_2 = test_util.generate_object_url(
        GCS_TESTING_PATH, OBJECT_NAME + "-shard2.tfrecord", test_id=test_id)
    logger.info(f"Writing dataset to {gcs_path_1} and {gcs_path_2}")
    with tf.io.TFRecordWriter(gcs_path_1) as writer:
        for idx, record in enumerate(serialized_dataset):
            if idx % 2 == 0:
                writer.write(record.numpy())
    with tf.io.TFRecordWriter(gcs_path_2) as writer:
        for idx, record in enumerate(serialized_dataset):
            if idx % 2 != 0:
                writer.write(record.numpy())

    # Reading dataset from GCS using tf.data.TFRecordDataaset
    logger.info(
        f"Reading dataset from {gcs_path_1} and {gcs_path_2} using tf.data.TFRecordDataaset")
    raw_dataset_1 = tf.data.TFRecordDataset(gcs_path_1)
    raw_dataset_2 = tf.data.TFRecordDataset(gcs_path_2)
    raw_dataset = raw_dataset_1.concatenate(raw_dataset_2)
    parsed_dataset = raw_dataset.map(test_util.parse_tfrecord_fn)

    actual = {"texts": []}
    for parsed_record in parsed_dataset:
        decoded_text = tf.io.decode_base64(parsed_record["text"])
        actual["texts"].append(decoded_text.numpy().decode('utf-8'))

    assert sorted(expected["texts"]) == sorted(actual["texts"])

    # Reading dataset from GCS using axlearn.tfrecord_dataset fn
    logger.info(
        f"Reading dataset from {gcs_path_1} and {gcs_path_2} using alxearn.tfrecord_dataset fn")
    pattern = gcs_path_1.replace(
        OBJECT_NAME + "-shard1.tfrecord", OBJECT_NAME + "*")
    source = config_for_function(tfrecord_dataset).set(
        glob_path=pattern,
        is_training=False,
        shuffle_buffer_size=0,
        features={
            # Single string (text) feature
            "text": tf.io.FixedLenFeature([], tf.string)
        }
    )
    ds = source.instantiate()

    actual = {"texts": []}
    for input_batch in ds():
        decoded_text = tf.io.decode_base64(input_batch["text"])
        actual["texts"].append(decoded_text.numpy().decode('utf-8'))

    assert sorted(expected["texts"]) == sorted(actual["texts"])


@pytest.mark.skip(reason="temp")
def test_tf_tensorstore_write_read_chunked(setup_data):
    """Test case for tensortore - read from GCS"""
    assert True


@pytest.mark.skip(reason="temp")
def test_tf_tensorstore_write_read_simple(setup_data):
    """Test case for tensortore - write to GCS. i.e orbax checkpoint."""
    from google.cloud import storage

    # Instantiate a client
    storage_client = storage.Client()
    storage_client.get_bucket("eshen-gcs-proxy-2")
    expected = setup_data["original_object"]
    object_url = test_util.generate_object_url(GCS_TESTING_PATH, OBJECT_NAME)

    # TensorStore specification using the JSON format
    spec = {
        "driver": "json",
        "kvstore": {
            "driver": "gcs",
            "bucket": TEST_BUCKET
        },
        "path": object_url.replace(f"gs://{TEST_BUCKET}/", "")
    }

    store = ts.open(spec, create=True, open=True).result()
    store.write(expected).result()

    actual = store.read().result()
    assert expected == actual


@pytest.mark.skip(reason="temp")
def test_tf_summary_write(setup_data):
    """Test case for tf.summary - write to GCS. i.e tf native checkpoint."""
    assert True


@pytest.mark.skip(reason="tfds is used to read public curated data from GCS. No need to encrypt.")
def test_tfds_write_read(setup_data):
    """Test case for tensorflow-dataset. Load public curated data from GCS"""
    assert True

# Add more test functions as needed


if __name__ == "__main__":
    pytest.main()
