package main

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

func HandleSimpleDownloadRequest(f *proxy.Flow) error {
	// handle streaming downloads in an ineffecient way. download whole file and return range.
	byteRangeHeader := f.Request.Header.Get("range")
	if byteRangeHeader != "" {
		f.Request.Header.Set("x-original-byte-range", byteRangeHeader)
		f.Request.Header.Del("range")
	}

	return nil
}

func HandleSimpleDownloadResponse(f *proxy.Flow) error {
	log.Debug(fmt.Sprintf("Got data in HandleSimpleDownloadResponse :%v", len(f.Response.Body)))

	log.Debug(fmt.Sprintf("Got data in HandleSimpleDownloadResponse %s", f.Response.Body))
	bucketName := getBucketNameFromRequestUri(f.Request.URL.Path)
	// Update the response content with the decrypted content
	unencryptedBytes, err := decryptBytes(f.Request.Raw().Context(),
		getKMSKeyName(bucketName), //config.KmsResourceName,
		f.Response.Body)
	if err != nil {
		return fmt.Errorf("unable to decrypt response body:%v", err)

	}

	log.Debug("#### Decryption OK")
	// check if this was as streaming/chunked download
	byteRangeHeader := f.Request.Header.Get("x-original-byte-range")

	if byteRangeHeader != "" {
		log.Debugf("Grabbing requested byte range slice %v", byteRangeHeader)
		start, end, _, err := parseByteRangeHeader(byteRangeHeader)
		if err != nil {
			return err
		}
		unencryptedByteSlice := unencryptedBytes[start:end]
		unencryptedBytes = unencryptedByteSlice //TODO: Performance/profiling
	}

	f.Response.Body = unencryptedBytes
	contentLength := bytes.Count(unencryptedBytes, []byte{})

	log.Debugf("#### Unencrypted Length: %v", contentLength)

	// Update content length headers with new length of decrypted data
	f.Response.Header.Set("X-Goog-Stored-Content-Length", strconv.Itoa(contentLength))
	f.Response.Header.Set("Content-Length", strconv.Itoa(contentLength))

	hashValue := base64_md5hash(unencryptedBytes)
	f.Response.Header.Set("X-Goog-Hash", hashValue)

	return nil

}
