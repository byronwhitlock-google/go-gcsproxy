package main

import (
	"bytes"
	"fmt"
	"strconv"

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
	log.Debug(fmt.Sprintf("Got data in HandleSimpleDownloadResponse %s", f.Response.Body))

	// Update the response content with the decrypted content
	unencryptedBytes, err := decryptBytes(f.Request.Raw().Context(),
		config.KmsResourceName, f.Response.Body)
	if err != nil {
		return fmt.Errorf("unable to decrypt response body:%v", err)

	}

	log.Debug("#### Decryption OK:")
	log.Debug(string(unencryptedBytes))

	f.Response.Body = unencryptedBytes
	contentLength := bytes.Count(unencryptedBytes, []byte{})

	log.Debug("#### Unencrypted Length:")
	fmt.Println(contentLength)

	// Update content length headers with new length of decrypted data
	f.Response.Header.Set("X-Goog-Stored-Content-Length", strconv.Itoa(contentLength))
	f.Response.Header.Set("Content-Length", strconv.Itoa(contentLength))

	hashValue := base64_md5hash(unencryptedBytes)
	f.Response.Header.Set("X-Goog-Hash", hashValue)

	return nil

}
