package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/google/tink/go/tink"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKMSClient is a mock implementation of the KMS client
type MockKMSClient struct {
    mock.Mock
}

func (m *MockKMSClient) Supported(string) bool {
	return true
}

// MockAEAD is a mock implementation of the AEAD interface
type MockAEAD struct {
    mock.Mock
}

func (m *MockAEAD) Encrypt(plaintext, additionalData []byte) ([]byte, error) {
    args := m.Called(plaintext, additionalData)
	fmt.Println(args)
    return args.Get(0).([]byte), args.Error(1)
}

func (m *MockAEAD) Decrypt(ciphertext, additionalData []byte) ([]byte, error) {
    args := m.Called(ciphertext, additionalData)
	fmt.Println(args)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockKMSClient) GetAEAD(uri string) (tink.AEAD, error) {
    args := m.Called(uri)
    return args.Get(0).(tink.AEAD), args.Error(1)
}


func TestEncryptBytesdo(t *testing.T) {
    tests := []struct {
        name          string
        resourceName  string
        inputBytes    []byte
        expectedBytes []byte
        mockSetup     func(*MockKMSClient, *MockAEAD)
        expectError   bool
        errorMessage  string
    }{
        {
            name:         "successful encryption",
            resourceName: "projects/test-project/locations/global/keyRings/test-ring/cryptoKeys/test-key",
            inputBytes:   []byte("test data"),
            expectedBytes: []byte("encrypted_test_data"),
            mockSetup: func(mockClient *MockKMSClient, mockAead *MockAEAD) {
                mockClient.On("GetAEAD", mock.Anything).Return(mockAead, nil)
                mockAead.On("Encrypt", mock.Anything, mock.Anything).Return([]byte("encrypted_test_data"), nil)
            },
            expectError: false,
        },
		{
            name:         "kms client creation fails",
            resourceName: "invalid-resource",
            inputBytes:   []byte("test data"),
            mockSetup: func(mockClient *MockKMSClient, mockAead *MockAEAD) {
                mockClient.On("GetAEAD", mock.Anything).Return(mockAead, nil)
                mockAead.On("Encrypt", mock.Anything, mock.Anything).Return([]byte(""), fmt.Errorf("Encryption error"))
            },
            expectError:  true,
			errorMessage: "error encrypting data: Encryption error",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            mockClient := new(MockKMSClient)
            mockAead := new(MockAEAD)
            tt.mockSetup(mockClient, mockAead)

            // Call the function
            result, err := doEncryptBytes(mockClient, tt.resourceName, tt.inputBytes)

            // Assert results
            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMessage)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
            }

            // Verify that all expected mock calls were made
            mockClient.AssertExpectations(t)
            mockAead.AssertExpectations(t)
        })
    }
}


func TestDecryptBytesdo(t *testing.T) {
    tests := []struct {
        name          string
        resourceName  string
        inputBytes    []byte
        expectedBytes []byte
        mockSetup     func(*MockKMSClient, *MockAEAD)
        expectError   bool
        errorMessage  string
    }{
        {
            name:         "successful encryption",
            resourceName: "projects/test-project/locations/global/keyRings/test-ring/cryptoKeys/test-key",
            inputBytes:   []byte("encrypted_test_data"),
            expectedBytes: []byte("test data"),
            mockSetup: func(mockClient *MockKMSClient, mockAead *MockAEAD) {
                mockClient.On("GetAEAD", mock.Anything).Return(mockAead, nil)
                mockAead.On("Encrypt", mock.Anything, mock.Anything).Return([]byte("encrypted_test_data"), nil)
            },
            expectError: false,
        },
		{
            name:         "kms client creation fails",
            resourceName: "invalid-resource",
            inputBytes:   []byte("encrypted_test_data"),
            mockSetup: func(mockClient *MockKMSClient, mockAead *MockAEAD) {
                mockClient.On("GetAEAD", mock.Anything).Return(mockAead, nil)
                mockAead.On("Encrypt", mock.Anything, mock.Anything).Return([]byte(""), fmt.Errorf("Encryption error"))
            },
            expectError:  true,
			errorMessage: "error encrypting data: Encryption error",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            mockClient := new(MockKMSClient)
            mockAead := new(MockAEAD)
            tt.mockSetup(mockClient, mockAead)

            // Call the function
            result, err := doDecryptBytes(mockClient, tt.resourceName, tt.inputBytes)

            // Assert results
            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMessage)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
            }

            // Verify that all expected mock calls were made
            mockClient.AssertExpectations(t)
            mockAead.AssertExpectations(t)
        })
    }
}


func TestBase64MD5Hash_Success(t *testing.T) {
	// Test case 1: A normal byte stream
	byteStream := []byte("test-string")
	expectedMD5 := md5.Sum(byteStream)
	expectedBase64MD5 := base64.StdEncoding.EncodeToString(expectedMD5[:])

	result := Base64MD5Hash(byteStream)
	assert.Equal(t, expectedBase64MD5, result, "Base64MD5Hash should return the correct hash")
}

func TestBase64MD5Hash_EmptyInput(t *testing.T) {
	// Test case 2: Empty byte stream
	byteStream := []byte{}
	expectedMD5 := md5.Sum(byteStream)
	expectedBase64MD5 := base64.StdEncoding.EncodeToString(expectedMD5[:])

	result := Base64MD5Hash(byteStream)
	assert.Equal(t, expectedBase64MD5, result, "Base64MD5Hash should return correct hash for empty input")
}

func TestBase64MD5Hash_NilInput(t *testing.T) {
	// Test case 3: Nil byte stream
	var byteStream []byte
	expectedMD5 := md5.Sum(byteStream)
	expectedBase64MD5 := base64.StdEncoding.EncodeToString(expectedMD5[:])

	result := Base64MD5Hash(byteStream)
	assert.Equal(t, expectedBase64MD5, result, "Base64MD5Hash should return correct hash for nil input")
}