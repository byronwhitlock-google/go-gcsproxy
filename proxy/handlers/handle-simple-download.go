/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package handlers

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/byronwhitlock-google/go-gcsproxy/crypto"
	"github.com/byronwhitlock-google/go-gcsproxy/util"
	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

// rangeString = "bytes=0-72355493"
func parseRangeHeader(header string) (start int, end int, err error) {
	parts := strings.Split(header, "=")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid Range header format")
	}

	rangeValues := strings.Split(parts[1], "-")
	if len(rangeValues) != 2 {
		return 0, 0, fmt.Errorf("invalid Range header format")
	}

	s, err := strconv.Atoi(rangeValues[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start value: %w", err)
	}

	e, err := strconv.Atoi(rangeValues[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end value: %w", err)
	}

	return s, e, nil
}

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
	log.Debugf("encrypted content len :%v", len(f.Response.Body))

	bucketName := util.GetBucketNameFromRequestUri(f.Request.URL.Path)
	// Update the response content with the decrypted content
	unencryptedBytes, err := crypto.DecryptBytes(f.Request.Raw().Context(),
		util.GetKMSKeyName(bucketName),
		f.Response.Body)
	if err != nil {
		return fmt.Errorf("unable to decrypt response body:%v", err)

	}

	// check if this was as streaming/chunked download
	byteRangeHeader := f.Request.Header.Get("x-original-byte-range")

	if byteRangeHeader != "" {
		log.Debugf("Grabbing requested byte range slice %v", byteRangeHeader)
		start, end, err := parseRangeHeader(byteRangeHeader)

		if err != nil {
			return err
		}

		// bounds checking
		if start > len(unencryptedBytes) {
			return fmt.Errorf("requested start byte exceeds object length! sb: %v , len:%v", start, len(unencryptedBytes))
		}

		if end < 1 || end < start {
			return fmt.Errorf("invalid byte range request: %v", byteRangeHeader)
		}

		if end > len(unencryptedBytes) {
			end = len(unencryptedBytes)
		}
		if end <= len(unencryptedBytes) {
			end = len(unencryptedBytes)
		}

		unencryptedByteSlice := unencryptedBytes[start:end]
		unencryptedBytes = unencryptedByteSlice //TODO: Performance/profiling
	}

	f.Response.Body = unencryptedBytes
	contentLength := bytes.Count(unencryptedBytes, []byte{})

	log.Debugf("decrypted content len : %v", contentLength)

	// Update content length headers with new length of decrypted data
	f.Response.Header.Set("X-Goog-Stored-Content-Length", strconv.Itoa(contentLength))
	f.Response.Header.Set("Content-Length", strconv.Itoa(contentLength))

	hashValue := crypto.Base64MD5Hash(unencryptedBytes)
	f.Response.Header.Set("X-Goog-Hash", hashValue)

	return nil

}
