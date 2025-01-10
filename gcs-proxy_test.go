package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
)

// MockResponseWriter is a mock implementation of http.ResponseWriter.
type MockResponseWriter struct {
	header http.Header
	body   *bytes.Buffer
	status int
}

func NewMockResponseWriter() *MockResponseWriter {
	return &MockResponseWriter{
		header: make(http.Header),
		body:   new(bytes.Buffer),
	}
}

func (m *MockResponseWriter) Header() http.Header {
	return m.header
}

func (m *MockResponseWriter) Write(b []byte) (int, error) {
	return m.body.Write(b)
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

// MockFlow simulates a proxy.Flow for testing purposes.
type MockFlow struct {
	proxy.Flow
	Request  *http.Request
	Response *http.Response
	Ctx      *MockContext // Add a context to hold client connection information
}

// MockClientConn simulates a net.Conn for testing.
type MockClientConn struct{}

// MockContext holds the client connection and response writer.
type MockContext struct {
	ClientConn     *MockClientConn
	ResponseWriter *MockResponseWriter
}

func (c *MockContext) ResponseWriter_() http.ResponseWriter {
	return c.ResponseWriter
}

func NewMockFlow(req *http.Request, resp *http.Response) *MockFlow {
	url, _ := url.Parse(req.URL.String())

	flow := &proxy.Flow{
		Request: &proxy.Request{
			URL: url,
		},
		ConnContext: &proxy.ConnContext{
			ClientConn: &proxy.ClientConn{},
		},
	}

	if resp != nil {
		flow.Response = &proxy.Response{
			StatusCode: 200,
			Header:     http.Header{},
			Body:       []byte{'t', 'e', 'e', 's', 't'},
		}
	} else {
		flow.Response = nil
	}

	mockCtx := &MockContext{
		ClientConn:     &MockClientConn{},
		ResponseWriter: NewMockResponseWriter(),
	}

	return &MockFlow{
		Flow:     *flow,
		Request:  req,
		Response: resp,
		Ctx:      mockCtx,
	}
}

func (mf *MockFlow) Raw() *http.Request {
	return mf.Request
}

func (mf *MockFlow) ReplaceToDecodedBody() {
	// Do nothing in the mock implementation
}

func TestInterceptGcsMethod(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		method   string
		expected gcsMethod
	}{
		{
			name:     "Multipart upload",
			url:      "https://storage.googleapis.com/upload/storage/v1/b/bucket/o?uploadType=multipart",
			method:   "POST",
			expected: multiPartUpload,
		},
		{
			name:     "Single part upload",
			url:      "https://storage.googleapis.com/upload/storage/v1/b/bucket/o?uploadType=media",
			method:   "POST",
			expected: singlePartUpload,
		},
		{
			name:     "Resumable upload post",
			url:      "https://storage.googleapis.com/resumable/upload/storage/v1/b/bucket/o",
			method:   "POST",
			expected: resumableUploadPost,
		},
		{
			name:     "Resumable upload put",
			url:      "https://storage.googleapis.com/upload/storage/v1/b/bucket/o",
			method:   "PUT",
			expected: resumableUploadPut,
		},
		{
			name:     "Simple download",
			url:      "https://storage.googleapis.com/download/storage/v1/b/bucket/o",
			method:   "GET",
			expected: simpleDownload,
		},
		{
			name:     "Metadata request",
			url:      "https://storage.googleapis.com/storage/v1/b/bucket/o?alt=json",
			method:   "GET",
			expected: metadataRequest,
		},
		{
			name:     "Pass through - different host",
			url:      "https://example.com/foo",
			method:   "GET",
			expected: passThru,
		},
		{
			name:     "Pass through - different path",
			url:      "https://storage.googleapis.com/foo",
			method:   "GET",
			expected: passThru,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, nil)
			flow := NewMockFlow(req, nil)
			result := InterceptGcsMethod(&flow.Flow)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestDecryptGcsPayloadResponse(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		method   string
		expected gcsMethod
	}{
		{
			name:     "Multipart upload",
			url:      "https://storage.googleapis.com/upload/storage/v1/b/bucket/o?uploadType=multipart",
			method:   "POST",
			expected: multiPartUpload,
		},
		{
			name:     "Simple download",
			url:      "https://storage.googleapis.com/download/storage/v1/b/bucket/o",
			method:   "GET",
			expected: simpleDownload,
		},
		// Add more test cases for other GCS methods
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, nil)

			// Create a mock response using httptest.ResponseRecorder
			recorder := httptest.NewRecorder()
			recorder.Body = bytes.NewBufferString("mock response body")
			recorder.Code = 200
			recorder.HeaderMap = make(http.Header)

			resp := recorder.Result()

			flow := NewMockFlow(req, resp)
			decryptGcsPayload := &DecryptGcsPayload{}
			decryptGcsPayload.Response(&flow.Flow)
		})
	}
}
