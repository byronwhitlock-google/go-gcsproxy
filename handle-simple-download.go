package main

import (
	"fmt"
	"strconv"

	"github.com/lqqyt2423/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

func HandleSimpleDownloadResponse(f *proxy.Flow) error {
	fmt.Println("simpleDownload")

	// Update the response content with the decrypted content
	unencryptedBytes, err := decryptBytes(f.Request.Raw().Context(),
		config.KmsResourceName, f.Response.Body)
	if err != nil {
		fmt.Println("Unable to decrypt response body:", err)
		log.Fatal(err)
	}

	fmt.Println(unencryptedBytes)
	fmt.Println(len(unencryptedBytes))
	f.Response.Body = unencryptedBytes
	contentLength := len(f.Response.Body)

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
