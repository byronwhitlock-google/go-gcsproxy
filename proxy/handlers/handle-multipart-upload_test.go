package handlers

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"testing"

	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	"github.com/stretchr/testify/assert"
)

// MockEncryptor is a mock implementation of the OtherPackagesInterface interface.
type MockOtherPackages struct {
	//mock.Mock
}

func (m *MockOtherPackages) GetCryptoEncryptBytes(ctx context.Context, kmsKeyName string, plaintext []byte) ([]byte, error) {
	//args := m.Called(ctx, kmsKeyName, plaintext)
	return []byte("encrypted-content"), nil
}

func (m *MockOtherPackages) GetCryptoBase64MD5Hash(plaintext []byte) (string) {
	return "i19y/5OSfdA3/sOa9Ml+Aw=="
}

func (m *MockOtherPackages) GetUtilGetBucketNameFromGcsMetadata(gcsmetadata map[string]interface{})(string) {
	return "test-bucket"
}

func (m *MockOtherPackages) GetUtilGetKMSKeyName(bucketName string)(string) {
	return "test-kms-key"
}


func TestHandleMultipartRequest(t *testing.T) {
	// Test cases for various scenarios
	testCases := []struct {
		name           string
		contentType    string
		requestBody    []byte
		expectedError  string
		//mockSetup      func(m *MockOtherPackages)
		expectedBody   []byte // Add this field
		kmsKeyName     string
	}{
		{
			name:        "Success",
			contentType: "multipart/related; boundary=\"foo\"",
			requestBody: []byte("--foo\r\nContent-Type: application/json\r\n\r\n{}\r\n--foo\r\nContent-Type: text/plain\r\n\r\nunencrypted-content\r\n--foo--"),
			// mockSetup: func(m *MockOtherPackages) {
			// 	//m.On("GetUtilGetBucketNameFromGcsMetadata", mock.Anything).Return("test-bucket")
			// 	//m.On("GetUtilGetKMSKeyName", "test-bucket").Return("test-kms-key")
			// 	//m.On("GetCryptoEncryptBytes", mock.Anything, "test-kms-key", []byte("unencrypted-content")).Return([]byte("encrypted-content"), nil)
			// 	//m.On("GetCryptoBase64MD5Hash", []byte("unencrypted-content")).Return("i19y/5OSfdA3/sOa9Ml+Aw==")

			// },
			expectedBody: []byte("--foo\r\nContent-Type: application/json\r\n\r\n{\"metadata\":{\"x-md5Hash\":\"i19y/5OSfdA3/sOa9Ml+Aw==\",\"x-unencrypted-content-length\":17}}\r\n--foo\r\nContent-Type: text/plain\r\n\r\nencrypted-content\r\n--foo--"),
			kmsKeyName:   "test-kms-key",
		},
		// {
		// 	name:        "InvalidContentType",
		// 	contentType: "invalid-content-type",
		// 	requestBody: []byte("--foo\r\nContent-Type: application/json\r\n\r\n{}\r\n--foo\r\nContent-Type: text/plain\r\n\r\nunencrypted-content\r\n--foo--"),
		// 	expectedError: "error parsing content type :",
		// },
		// {
		// 	name:        "EncryptError",
		// 	contentType: "multipart/related; boundary=\"foo\"",
		// 	requestBody: []byte("--foo\r\nContent-Type: application/json\r\n\r\n{}\r\n--foo\r\nContent-Type: text/plain\r\n\r\nunencrypted-content\r\n--foo--"),
		// 	mockSetup: func(m *MockOtherPackages) {
		// 		m.On("Encrypt", mock.Anything, "test-kms-key", []byte("unencrypted-content")).Return(nil, fmt.Errorf("encryption error"))
		// 	},
		// 	expectedError: "error encrypting request:",
		// 	kmsKeyName:   "test-kms-key",
		// },
	}


	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//mockOtherPackages := new(MockOtherPackages)
			mockOtherPackages := &MockOtherPackages{}
			// if tc.mockSetup != nil {
			// 	tc.mockSetup(mockOtherPackages)
			// }
			

			f := &proxy.Flow{
				Request: &proxy.Request{
					Header: http.Header{"Content-Type": []string{tc.contentType}},
					Body:   tc.requestBody,
				},
			}

			err := HandleMultipartRequestWrapper(f, mockOtherPackages)
			fmt.Println(err)

			// if tc.expectedError != "" {
			// 	assert.ErrorContains(t, err, tc.expectedError)
			// } else {
			// 	assert.NoError(t, err)
			// 	assert.Equal(t, tc.expectedBody, f.Request.Body)
			// }
			// mockOtherPackages.AssertExpectations(t)

		})
	}
}

// Test helper functions
func TestGetMultipartMimeHeader(t *testing.T) {
	part := &multipart.Part{
		Header: textproto.MIMEHeader{
			"Content-Type": {"application/json"},
			"X-Test-Header": {"test-value"},
		},
	}

	expected := textproto.MIMEHeader{
		"Content-Type": {"application/json"},
		"X-Test-Header": {"test-value"},
	}

	actual := GetMultipartMimeHeader(part)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Headers mismatch. Expected: %v, Got: %v", expected, actual)
	}
}


func TestGetMultipartMimeHeaderOctetStream(t *testing.T) {
	header := GetMultipartMimeHeaderOctetStream()
	expectedHeader := textproto.MIMEHeader{
		"Content-Type": {"application/octet-stream"},
	}
	if !reflect.DeepEqual(header, expectedHeader) {
		t.Errorf("Header mismatch. Expected: %v, Got: %v", expectedHeader, header)
	}
}

func TestHandleMultipartResponse(t *testing.T) {

	testCases := []struct {
		name           string
		responseBody   []byte
		requestHeaders http.Header
		expectedBody   []byte
		expectedError  string
	}{
		{
			name:         "Success",
			responseBody: []byte(`{"kind": "storage#object", "md5Hash": "old-hash", "size": "123"}`),
			requestHeaders: map[string][]string{
				"gcs-proxy-original-md5-hash":        []string{"new-hash"},
				"gcs-proxy-unencrypted-file-size": []string{"456"},
			},
			expectedBody: []byte(`{"kind": "storage#object", "md5Hash": "new-hash", "size": 456}`),
		},
		{
			name:           "InvalidJSON",
			responseBody:   []byte(`invalid-json`),
			expectedError:  "error unmarshalling JSON",
		},
		{
			name:         "MissingHeaders",
			responseBody: []byte(`{"kind": "storage#object", "md5Hash": "old-hash", "size": "123"}`),
			requestHeaders: map[string][]string{
				"gcs-proxy-original-md5-hash": []string{"new-hash"},
				// Missing gcs-proxy-unencrypted-file-size
			},
			expectedError: "error setting json response", // Expecting strconv.Atoi error
		},
		{
			name: "InvalidSizeHeader",
			responseBody: []byte(`{"kind": "storage#object", "md5Hash": "old-hash", "size": "123"}`),
			requestHeaders: map[string][]string{
				"gcs-proxy-original-md5-hash":        []string{"new-hash"},
				"gcs-proxy-unencrypted-file-size": []string{"invalid-size"},
			},
			expectedError: "error setting json response", // Expecting strconv.Atoi error

		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			headers := http.Header{}
			for k, v := range tc.requestHeaders {
				headers.Set(k, v[0])
			}

			f := &proxy.Flow{
				Request: &proxy.Request{
					Header: headers,
				},
				Response: &proxy.Response{
					Body: tc.responseBody,
				},
			}

			err := HandleMultipartResponse(f)

			if tc.expectedError != "" {
				assert.ErrorContains(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.JSONEq(t, string(tc.expectedBody), string(f.Response.Body))
			}
		})
	}
}




