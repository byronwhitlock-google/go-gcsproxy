"""
Proxy funtioanl testing for Axlearn(tensforflow) libraries to interface GCS

Setup:
  Set the following the enviroment variables:
   -- PROXY_FUNC_TEST_BUCKET: GCS bucket for testing. Required
   -- https_proxy: Point to the proxy. Required
                   ie. https_proxy=https://localhost:8080
   -- SSL_CERT_FILE: Mitmproxy self-signed ca cert. For orbax/tensorstore. Required
   -- CURL_CA_BUNDLE: Mitmproxy self-signed ca cert. For tf.io and tf.data. However, I couldn't 
                    make it work and had to add the cert to the system store.
   -- GCS_RETRY_CONFIG_MAX_RETRIES: Control GCS retry for tf.io and tf.data. Default is 10. Set 
                    it to 0 to disable retry. Optional.
   
   If needed, add the mitmproxy self-signed ca cert to the system store: 
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
from axlearn.common import file_system as fs
from axlearn.common.config import config_for_function
from axlearn.common.input_tf_data import (tfrecord_dataset)
import tensorstore as ts
import jax
import jax.numpy as jnp

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


def test_axlearn_fileio_copy(setup_data):
    """Test case for axlearn file_io.copy()

       Axlearn regular [checkpointer](https://github.com/apple/axlearn/blob/main/axlearn/common/checkpointer.py)
       uses copy() to write checkpoints to GCS.

    """
    test_id = test_axlearn_fileio_copy.__name__
    source = "/tmp/source"
    object_url = test_util.generate_object_url(
        GCS_TESTING_PATH, OBJECT_NAME, test_id=test_id)

    expected = setup_data["original_object"]
    with open(source, 'w') as file:
        file.write(expected)

    logger.info(f"Copying {source} to {object_url}")
    fs.copy(source, object_url, overwrite=True)

    logger.info(f"Getting {object_url}")
    with fs.open(object_url) as f:
        actual = f.read()
    assert expected == actual


def test_tf_io_gfile_write_read(setup_data):
    """Test case for tf.io.gfile.GFile.write()"""
    test_id = test_tf_io_gfile_write_read.__name__
    expected = setup_data["original_object"]
    object_url = test_util.generate_object_url(
        GCS_TESTING_PATH, OBJECT_NAME, test_id=test_id)

    logger.info(f"Creating {object_url}")
    with tf.io.gfile.GFile(object_url, "w") as f:
        f.write(expected)

    logger.info(f"Getting {object_url}")
    with tf.io.gfile.GFile(object_url, "r") as f:
        actual = f.read()
    assert expected == actual


def test_tf_data_write_read(setup_data):
    """Test case for tf.data.TFRecordDataset which is used by axlearn input_tf_data.tfrecrod_dataset"""
    test_id = test_tf_data_write_read.__name__
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

#uncomment to skip if gcs proxy encryption is enabled
#@pytest.mark.skip(reason="Not working with encryption yet.")
def test_tensorstore_orbax_write_read_pytree(setup_data):
    """Test case for orbax/tensortore - write jax pytree to GCS with ocdbt driver which orbax uses."""
    test_id = test_tensorstore_orbax_write_read_pytree.__name__
    expected = {"a": jnp.array([1, 2, 3]), "b": jnp.ones((2, 2))}
    bucket_name = TEST_BUCKET
    object_url = test_util.generate_object_url(
        GCS_TESTING_PATH, object_name="", test_id=test_id)
    object_prefix = object_url.rstrip('/').replace(f"gs://{TEST_BUCKET}/", "")

    leaves, tree_def = jax.tree_util.tree_flatten(expected)
    logger.info(f"Writing PyTree data to {object_url}")
    leaf_paths = []
    for i, leaf in enumerate(leaves):
        path = f"{object_prefix}/leaf_{i}"
        spec = {
            "driver": "zarr",  # Use the Zarr driver for storing tensor data
            "kvstore": {  # Specify OCDBT as the kvstore
                "driver": "ocdbt",  # OCDBT as the key-value store
                "base": {  # Base storage is GCS
                    "driver": "gcs",
                    "bucket": bucket_name,  # Specify the GCS bucket
                    "path": path
                }
            },
            "path": path
        }

        # Save the leaf
        ts_array = ts.open(
            spec,
            create=True,
            dtype=leaf.dtype,
            shape=leaf.shape
        ).result()
        ts_array[...] = jax.device_get(leaf)
        leaf_paths.append(path)

    logger.info(f"Getting PyTree data to {object_url}")
    actual_leaves = []
    for path in leaf_paths:
        spec = {
            "driver": "zarr",  # Use the Zarr driver for storing tensor data
            "kvstore": {  # Specify OCDBT as the kvstore
                "driver": "ocdbt",  # OCDBT as the key-value store
                "base": {  # Base storage is GCS
                    "driver": "gcs",
                    "bucket": bucket_name,  # Specify the GCS bucket
                    "path": path
                }
            },
            "path": path
        }
        ts_array = ts.open(spec).result()
        actual_leaves.append(jnp.array(ts_array[...]))

    # Load PyTree
    actual = jax.tree_util.tree_unflatten(tree_def, actual_leaves)

    assert actual.keys() == expected.keys()
    for key in actual:
        assert jnp.array_equal(actual[key], expected[key])


def test_tf_tensorstore_write_read_simple(setup_data):
    """Test case for tensortore - write to GCS with single file driver."""
    test_id = test_tf_tensorstore_write_read_simple.__name__
    expected = setup_data["original_object"]
    object_url = test_util.generate_object_url(
        GCS_TESTING_PATH, OBJECT_NAME, test_id=test_id)

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
    logger.info(f"Writing object to {object_url}")
    store.write(expected).result()
    logger.info(f"Reading object from {object_url}")
    actual = store.read().result()
    assert expected == actual


@pytest.mark.skip(reason="tfds is used to read public curated data from GCS. No need to encrypt.")
def test_tfds_write_read(setup_data):
    """Test case for tensorflow-dataset. Load public curated data from GCS"""
    assert True

# Add more test functions as needed


if __name__ == "__main__":
    pytest.main()
