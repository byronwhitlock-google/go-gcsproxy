import logging
import uuid
from typing import Sequence
import tensorflow as tf

logger = logging.getLogger(__name__)


def generate_object_url(gcs_testing_path: str, object_name: str, test_id: str = None, object_suffix: str = None) -> str:
    test_id_part = test_id if test_id else str(uuid.uuid4())

    suffix_part = f"-{object_suffix}" if object_suffix else ""

    return f"{gcs_testing_path}/{test_id_part}/{object_name}{suffix_part}"


def build_ds_fn(texts: Sequence[str]):
    # del is_training, data_dir

    def ds_fn() -> tf.data.Dataset:
        def data_gen():
            yield from texts

        ds = tf.data.Dataset.from_generator(
            data_gen,
            output_signature=tf.TensorSpec(shape=(), dtype=tf.string)
        )
        ds = ds.apply(tf.data.experimental.assert_cardinality(len(texts)))
        return ds

    return ds_fn


def serialize_example(text):
    feature = {
        "text": tf.train.Feature(
            bytes_list=tf.train.BytesList(
                value=[tf.io.encode_base64(text).numpy()])
        )
    }
    example_proto = tf.train.Example(
        features=tf.train.Features(feature=feature))
    return example_proto.SerializeToString()


def tf_serialize_example(text):
    tf_string = tf.py_function(
        func=serialize_example,
        inp=[text],
        Tout=tf.string
    )
    return tf_string


def parse_tfrecord_fn(example):
    feature_description = {
        "text": tf.io.FixedLenFeature([], tf.string),
    }
    return tf.io.parse_single_example(example, feature_description)
