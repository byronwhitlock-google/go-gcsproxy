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
#from axlearn.common import input_tf_data as tf_data
from axlearn.common.config import config_for_function
from axlearn.common.input_tf_data import (tfds_dataset, tfrecord_dataset)

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


# def test_axlearn_fileio_copy(setup_data):
#     """Test case for axlearn file_io.copy()"""
#     from axlearn.common import file_system as fs
#     source = "/tmp/source"
#     object_url = test_util.generate_object_url(GCS_TESTING_PATH, OBJECT_NAME)
    
#     expected_content = setup_data["original_object"]
#     with open(source, 'w') as file:
#       file.write(expected_content)
      
#     logger.info(f"Copying {source} to {object_url}")
#     fs.copy(source, object_url, overwrite=True)
    
#     logger.info(f"Getting {object_url}")   
#     with fs.open(object_url) as f:
#         actual = f.read()
#     assert expected_content == actual

# def test_tf_io_gfile_write_read(setup_data):
#     """Test case for tf.io.gfile.GFile.write()"""
#     expected_content = setup_data["original_object"]
#     object_url = test_util.generate_object_url(GCS_TESTING_PATH, OBJECT_NAME)
    
#     logger.info(f"Creating {object_url}")
#     with tf.io.gfile.GFile(object_url, "w") as f:
#         f.write(expected_content)
        
#     logger.info(f"Getting {object_url}")
#     with tf.io.gfile.GFile(object_url, "r") as f:
#         actual = f.read()
#     assert expected_content == actual

def test_tf_data_write_read(setup_data):
    """Test case for tf.data.TFRecordDataset which is used by axlearn input_tf_data.tfrecrod_dataset"""
    expected = {
        "texts": ["a", "b", "c", "d"]
    }

    # Create Dataset
    ds_fn = test_util.build_ds_fn(texts=expected["texts"])
    dataset = ds_fn()

    # Serialize Dataset
    serialized_dataset = dataset.map(test_util.tf_serialize_example)

    # Write to GCS in TFRecord Format (Sharded into two files)
    gcs_path_1 = test_util.generate_object_url(GCS_TESTING_PATH, OBJECT_NAME + "-shard1.tfrecord")
    gcs_path_2 = test_util.generate_object_url(GCS_TESTING_PATH, OBJECT_NAME + "-shard2.tfrecord")
    with tf.io.TFRecordWriter(gcs_path_1) as writer:
        for idx, record in enumerate(serialized_dataset):
            if idx % 2 == 0:
                writer.write(record.numpy())
    with tf.io.TFRecordWriter(gcs_path_2) as writer:
        for idx, record in enumerate(serialized_dataset):
            if idx % 2 != 0:
                writer.write(record.numpy())

    # Reading both TFRecord files from GCS
    logger.info(f"Reading dataset from {gcs_path_1} and {gcs_path_2}")
    raw_dataset_1 = tf.data.TFRecordDataset(gcs_path_1)
    raw_dataset_2 = tf.data.TFRecordDataset(gcs_path_2)

    # Combine both datasets
    raw_dataset = raw_dataset_1.concatenate(raw_dataset_2)
    parsed_dataset = raw_dataset.map(test_util.parse_tfrecord_fn)

    actual = {"texts": []}
    for parsed_record in parsed_dataset:
        decoded_text = tf.io.decode_base64(parsed_record["text"])
        actual["texts"].append(decoded_text.numpy().decode('utf-8'))

    # Assert that the data read from GCS matches the original data
    assert sorted(expected["texts"]) == sorted(actual["texts"])
    
    source = config_for_function(tfrecord_dataset).set(
        glob_path=gcs_path_1.replace(OBJECT_NAME + "-shard1.tfrecord", ""),
        is_training=False,
        shuffle_buffer_size=0,
        features={
            # Single string (text) feature
            "text": tf.io.FixedLenFeature([], tf.string)
        }
    )
    ds = source.instantiate()
    logger.info(ds)

@pytest.mark.skip(reason="temp")    
def test_eshen(setup_data):
    """
    eshen test
    """
    import tensorflow as tf

    # Step 1: Prepare Data
    data = {
        "features": [[1.0, 2.0], [3.0, 4.0], [5.0, 6.0]],
        "labels": [0, 1, 1]
    }
    
    dataset = tf.data.Dataset.from_tensor_slices((data["features"], data["labels"]))

    # Step 2: Serialize Function
    def serialize_example(features, label):
        feature = {
            "features": tf.train.Feature(float_list=tf.train.FloatList(value=features)),
            "label": tf.train.Feature(int64_list=tf.train.Int64List(value=[label]))
        }
        example_proto = tf.train.Example(features=tf.train.Features(feature=feature))
        return example_proto.SerializeToString()

    # Wrap with tf.py_function for Dataset
    def tf_serialize_example(features, label):
        tf_string = tf.py_function(
            func=serialize_example,
            inp=[features, label],
            Tout=tf.string
        )
        return tf_string

    # Step 3: Serialize Dataset
    serialized_dataset = dataset.map(tf_serialize_example)

    # Step 4: Write to GCS in TFRecord Format
    gcs_path = test_util.generate_object_url(GCS_TESTING_PATH, OBJECT_NAME)
    writer = tf.io.TFRecordWriter(gcs_path)
    logger.info(f"Writing dataset to {gcs_path}")
    with tf.io.TFRecordWriter(gcs_path) as writer:
        for record in serialized_dataset:
            writer.write(record.numpy())
            
    # Step 5: Read Back Dataset from GCS
    def parse_tfrecord_fn(example):
        feature_description = {
            "features": tf.io.FixedLenFeature([2], tf.float32),
            "label": tf.io.FixedLenFeature([], tf.int64),
        }
        return tf.io.parse_single_example(example, feature_description)

    # Read the TFRecord file from GCS
    logger.info(f"Reading dataset from {gcs_path}")
    raw_dataset = tf.data.TFRecordDataset(gcs_path)
    parsed_dataset = raw_dataset.map(parse_tfrecord_fn)

    # Verify the parsed data
    print("Reading back data from GCS:")
    for parsed_record in parsed_dataset:
        print(parsed_record)
        
    gcs_data = {"features": [], "labels": []}
    for parsed_record in parsed_dataset:
        gcs_data["features"].append(parsed_record["features"].numpy().tolist())
        gcs_data["labels"].append(parsed_record["label"].numpy())
    assert gcs_data == data



# def test_tf_data_write_read(setup_data):
#     """Test case for tf.data.TFRecordDataset which is used by axlearn input_tf_data.tfrecrod_dataset"""
#     #import tensorflow as tf
#     object_url = test_util.generate_object_url(GCS_TESTING_PATH, OBJECT_NAME)
#     # --- Writing to GCS ---

#     # Sample data (array of strings)
#     strings = ["apple", "banana", "cherry", "date"]
#     data = tf.data.Dataset.from_tensor_slices(strings)

#     # Convert data to TFRecord format (encoding strings as bytes)
#     # def serialize_example(s):
#     #     features = {
#     #         'text': tf.train.Feature(bytes_list=tf.train.BytesList(value=[s.encode('utf-8')]))
#     #     }
#     #     example_proto = tf.train.Example(features=tf.train.Features(feature=features))
#     #     return example_proto.SerializeToString()
#     # def serialize_example(s):
#     #     s = tf.ensure_tensor(s, tf.string)  # Ensure s is a tf.Tensor
#     #     features = {
#     #         'text': tf.train.Feature(bytes_list=tf.train.BytesList(value=[tf.strings.unicode_encode(s, 'utf-8')]))
#     #     }
#     #     example_proto = tf.train.Example(features=tf.train.Features(feature=features))
#     #     return example_proto.SerializeToString()

    
#     # def serialize_example(s):
#     #     # No need for tf.ensure_tensor in this case
#     #     features = {
#     #         'text': tf.train.Feature(bytes_list=tf.train.BytesList(value=[tf.compat.as_bytes(s)]))
#     #     }
#     #     example_proto = tf.train.Example(features=tf.train.Features(feature=features))
#     #     return example_proto.SerializeToString()
    
#     def serialize_example(s):
#         features = {
#             'text': tf.train.Feature(bytes_list=tf.train.BytesList(value=[tf.compat.as_bytes(s.numpy())]))  # Convert tensor to bytes
#         }
#         example_proto = tf.train.Example(features=tf.train.Features(feature=features))
#         return example_proto.SerializeToString()

#     data = data.map(serialize_example)

#     # Write to GCS
#     #output_path = "gs://your-bucket/your-strings.tfrecord"  # Replace with your GCS path
#     logger.info(f"Writing dataset to {object_url}")
#     writer = tf.data.experimental.TFRecordWriter(object_url)
#     writer.write(data)


#     # --- Reading from GCS ---

#     # Read from GCS
#     #input_path = "gs://your-bucket/your-strings.tfrecord"  # Replace with your GCS path
#     logger.info(f"Reading dataset from {object_url}")
#     dataset = tf.data.TFRecordDataset(object_url)

#     # Parse TFRecord data (decoding bytes back to strings)
#     def parse_example(serialized_example):
#         features = {
#             'text': tf.io.FixedLenFeature([], tf.string)
#         }
#         example = tf.io.parse_single_example(serialized_example, features)
#         return example['text']

#     dataset = dataset.map(parse_example)

#     # Iterate over the dataset
#     for element in dataset:
#         print(element)  # Output: tf.Tensor(b'apple', shape=(), dtype=string), etc.
#     assert True

def test_tf_tensorstore_read(setup_data):
    """Test case for tensortore - read from GCS"""
    assert True

def test_tf_tensorstore_write(setup_data):
    """Test case for tensortore - write to GCS. i.e orbax checkpoint."""
    assert True

def test_tf_summary_write(setup_data):
    """Test case for tf.summary - write to GCS. i.e tf native checkpoint."""
    assert True

def test_tfds_read(setup_data):
    """Test case for tensorflow-dataset. Load public curated data from GCS"""
    assert True

# Add more test functions as needed

if __name__ == "__main__":
    pytest.main()
