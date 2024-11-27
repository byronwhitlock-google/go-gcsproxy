package main

import (
	"fmt"
	"strconv"

	"github.com/lqqyt2423/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

func HandleSimpleDownloadResponse(f *proxy.Flow) error {
	log.Debug(fmt.Sprintf("Got data in HandleSimpleDownloadResponse %s", f.Response.Body))

	// Update the response content with the decrypted content
	unencryptedBytes, err := decryptBytes(f.Request.Raw().Context(),
		config.KmsResourceName, f.Response.Body)
	if err != nil {
		return fmt.Errorf("unable to decrypt response body:", err)

	}

	fmt.Println(unencryptedBytes)
	fmt.Println(len(unencryptedBytes))
	f.Response.Body = unencryptedBytes
	contentLength := len(unencryptedBytes)

	// Update content length headers with new length of decrypted data
	f.Response.Header.Set("X-Goog-Stored-Content-Length", strconv.Itoa(contentLength))
	f.Response.Header.Set("Content-Length", strconv.Itoa(contentLength))

	// gcloud storage cp command uses "range" in request
	//
	contentRange := "bytes 0-" + strconv.Itoa(contentLength-1) + "/" + strconv.Itoa(contentLength)

	f.Response.Header.Set("Content-Range", contentRange)

	hashValue := base64_md5hash(unencryptedBytes)
	f.Response.Header.Set("X-Goog-Hash", hashValue)

	return nil

}
