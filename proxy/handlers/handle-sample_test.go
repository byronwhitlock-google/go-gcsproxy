package handlers

import (
	"context"
	"testing"

	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

// MockCryptoClient is a mock implementation of the CryptoClient interface
type MockCryptoClient struct{}

// DecryptBytes mock implementation
func (m *MockCryptoClient) DecryptBytes(ctx context.Context, resourceName string, bytesToDecrypt []byte) ([]byte, error) {
	// Mock the decryption logic
	if resourceName == "key" && string(bytesToDecrypt) == "encryption" {
		return []byte("decrypted data"), nil
	}
	return nil, nil
}

func TestHandleSample(t *testing.T) {
	// Create a mock implementation of CryptoClient
	mockCryptoClient := &MockCryptoClient{}

	// Create a mock proxy.Flow
	request := &proxy.Request{}
	response := &proxy.Response{
		Body: []byte("encryption"),
	}
	mockFlow := &proxy.Flow{
		Request:  request,
		Response: response,
	}

	// Call the HandleSample function
	result := HandleSample(mockFlow, mockCryptoClient)

	// Verify the results
	assert.Equal(t, []byte("decrypted data"), mockFlow.Response.Body, "Response body should be updated with decrypted data")
	assert.Equal(t, 14, result, "Returned length should match the length of the decrypted data")
}
