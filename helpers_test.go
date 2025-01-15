package main

import (
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
)

//kms key name test
func TestGetKMSKeyName(t *testing.T) {
	config = &Config{KmsBucketKeyMapping: "test-bucket:test-key"}
	cases := []struct {
		bucketName  string
		kmsKeyName string
	}{
		{"test-bucket", "test-key"},
		{"test-bucket-1", ""},
		{"*", ""},
	}

	for _, c := range cases {
		kmsKeyName := getKMSKeyName(c.bucketName)
		if kmsKeyName != c.kmsKeyName {
			t.Errorf("getKMSKeyName(%q) == %q, want %q", c.bucketName, kmsKeyName, c.kmsKeyName)
		}
	}
}

func TestGetBucketNameFromGcsMetadata(t *testing.T) {
	testCases := []struct {
		input    map[string]interface{}
		expected string
	}{
		{map[string]interface{}{"name": "my-object"}, ""}, // Case where "bucket" key is missing
		{map[string]interface{}{}, ""},                   // Empty input map
	}

	for _, tc := range testCases {
		actual := getBucketNameFromGcsMetadata(tc.input)
		if actual != tc.expected {
			t.Errorf("For input %v, expected %q but got %q", tc.input, tc.expected, actual)
		}
	}
}



func TestGenerateHeadersList(t *testing.T) {
	f := &proxy.Flow{
		Request: &proxy.Request{
			Header: make(http.Header),
			Body:   []byte("test body"),
		},
	}

	headersMap, boundary := generateHeadersList(f)
	
	expectedHeaders := map[string]string{
		"Accept-Encoding": "gzip, deflate",
		"Accept":          "application/json",
		"Connection":      "keep-alive",
		"Content-Type":    "multipart/related; boundary='" + boundary + "'",
		"X-Goog-Api-Client": "cred-type/u",
		"Content-Length":strconv.Itoa(len(f.Request.Body)),
	}
	
	if !reflect.DeepEqual(headersMap, expectedHeaders) {
		t.Errorf("Headers mismatch. Expected: %v, Got: %v", expectedHeaders, headersMap)
	}

	if !strings.HasPrefix(boundary, "===============") || len(boundary) <= 14 {
		t.Errorf("Invalid boundary format: %s", boundary)
	}
}

func TestGetBucketNameFromRequestUri(t *testing.T) {
	testCases := []struct {
		urlPath  string
		expected string
	}{
		{"/download/storage/v1/b/ehorning-axlearn/o/README.md", "ehorning-axlearn"},
		{"/storage/v1/b/my-bucket/o/my-object", "my-bucket"},
		{"/upload/storage/v1/b/another-bucket/o", "another-bucket"},
		//{"/invalid/path", ""},
		{"/download/storage/v1/b/bucket/path/to/object/o/README.md","bucket"},
	}

	for _, tc := range testCases {
		actual := getBucketNameFromRequestUri(tc.urlPath)
		if actual != tc.expected {
			t.Errorf("For URL path %q, expected bucket name %q but got %q", tc.urlPath, tc.expected, actual)
		}
	}
}



func TestGenerateMetadata(t *testing.T) {

	f := &proxy.Flow{
		Request: &proxy.Request{
			URL: &url.URL{
				Path: "/upload/storage/v1/b/apple-lk-test2/o",
			},
			Header: make(http.Header),
			Body:   []byte("test body"),
		},
	}

	contentType := "text/plain"
	objectName := "test-object"
	metadata := generateMetadata(f, contentType, objectName)

	expectedMetadata := map[string]interface{}{
		"bucket":      "apple-lk-test2",
		"contentType": "text/plain",
		"name":        "test-object",
		"metadata": map[string]interface{}{
			"x-unencrypted-content-length": len("test body"),
			"x-md5Hash":                    Base64MD5Hash([]byte("test body")),
		},
	}

	if !reflect.DeepEqual(metadata, expectedMetadata) {
		t.Errorf("Metadata mismatch.\nExpected: %+v\nGot: %+v", expectedMetadata, metadata)
	}
}

func TestCreateFirstMultipartMimeHeader(t *testing.T) {
	header := CreateFirstMultipartMimeHeader()

	expectedHeader := textproto.MIMEHeader{
		"Content-Type": {"application/json"},
		"Mime-Version": {"1.0"},
	}

	if !reflect.DeepEqual(header, expectedHeader) {
		t.Errorf("Header mismatch. Expected: %v, Got: %v", expectedHeader, header)
	}
}

func TestCreateSecondMultipartMimeHeader(t *testing.T) {
	contentType := "text/plain"
	header := CreateSecondMultipartMimeHeader(contentType)
	expectedHeader := textproto.MIMEHeader{
		"Content-Type":           {contentType},
		"Mime-Version":           {"1.0"},
		"Content-Transfer-Encoding": {"binary"},
	}
	if !reflect.DeepEqual(header, expectedHeader) {
		t.Errorf("Header mismatch. Expected: %v, Got: %v", expectedHeader, header)
	}
}

func TestGenerateRandom19DigitNumber(t *testing.T) {
	num := generateRandom19DigitNumber()
	numStr := strconv.Itoa(num)

	if len(numStr) != 19 {
		t.Errorf("Generated number is not 19 digits long: %s", numStr)
	}

	if _, err := strconv.Atoi(numStr); err != nil {
		t.Errorf("Generated number is not a valid integer: %s", numStr)
	}
}


type mockPart struct {
	header textproto.MIMEHeader
	r io.Reader

}

func (m *mockPart) FormName() string {
	return "file"
}
func (m *mockPart) FileName() string {
	return "example.txt"
}
func (m *mockPart) Close() error {
	return nil
}
func (m *mockPart) Read(p []byte) (n int, err error) {
	return m.r.Read(p)
}




