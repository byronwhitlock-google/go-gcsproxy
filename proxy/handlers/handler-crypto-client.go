package handlers

import (
	"context"

	"github.com/byronwhitlock-google/go-gcsproxy/crypto"
)

type CryptoClient interface {
	DecryptBytes(ctx context.Context, resourceName string, bytesToDecrypt []byte) ([]byte, error)
}

type CryptoClientImpl struct {
}

func (c *CryptoClientImpl) DecryptBytes(ctx context.Context, resourceName string, bytesToDecrypt []byte) ([]byte, error) {
	return crypto.EncryptBytes(ctx, resourceName, bytesToDecrypt)
}
