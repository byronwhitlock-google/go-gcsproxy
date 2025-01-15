package main

import (
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
	//return []byte("encrypted_test_data"), nil
}

func (m *MockAEAD) Decrypt(ciphertext, additionalData []byte) ([]byte, error) {
    args := m.Called(ciphertext, additionalData)
	fmt.Println(args)
	return args.Get(0).([]byte), args.Error(1)
	//fmt.Println(args)
    //return []byte("test data"), nil
}

func (m *MockKMSClient) GetAEAD(uri string) (tink.AEAD, error) {
    args := m.Called(uri)
    return args.Get(0).(tink.AEAD), args.Error(1)
}


func TestEncryptBytes(t *testing.T) {
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
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            mockClient := new(MockKMSClient)
            mockAead := new(MockAEAD)
            tt.mockSetup(mockClient, mockAead)

            // Call the function
            result, err := EncryptBytes(mockClient, tt.resourceName, tt.inputBytes)

            // Assert results
            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMessage)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
                //assert.True(t, bytes.Equal(tt.expectedBytes, result))
            }

            // Verify that all expected mock calls were made
            mockClient.AssertExpectations(t)
            mockAead.AssertExpectations(t)
        })
    }
}
