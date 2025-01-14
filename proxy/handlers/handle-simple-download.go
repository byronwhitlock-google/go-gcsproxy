/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package handlers

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/byronwhitlock-google/go-gcsproxy/crypto"
	"github.com/byronwhitlock-google/go-gcsproxy/util"
	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

func HandleSimpleDownloadRequest(f *proxy.Flow) error {
	// update request grab the whole file.
	// strip range header, we don't support partial uploads at all.
	f.Request.Header.Del("range")
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
