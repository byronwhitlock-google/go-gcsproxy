import logging
import os
import tink
from tink import aead
from tink.integration import gcpkms


logger = logging.getLogger(__name__)


def get_aead(gcp_kms_key) -> aead.KmsEnvelopeAead:
    try:
        aead.register()
        client = gcpkms.GcpKmsClient(gcp_kms_key, None)

        remote_aead = client.get_aead(gcp_kms_key)
        env_aead = aead.KmsEnvelopeAead(
            aead.aead_key_templates.AES256_GCM, remote_aead
        )
        return env_aead
    except tink.TinkError as e:
        logger.error("Error getting aead: %s", e)
        #raise
