package main

import (
	"fmt"
	"strings"

	"github.com/lqqyt2423/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

type DecryptGcsPayload struct {
	proxy.BaseAddon
}
type EncryptGcsPayload struct {
	proxy.BaseAddon
}

// https://cloud.google.com/storage/docs/json_api/v1/objects
type gcsMethod int

const (
	multiPartUpload   gcsMethod = iota // uploadType=multipart, VERB=POST, uri=/upload/storage/v1/b/  DOCS: https://cloud.google.com/storage/docs/json_api/v1/objects/insert
	singlePartUpload                   // uploadType=media,     VERB=POST, uri=/upload/storage/v1/b/
	resumableUpload                    // unsupported uploadType=resumable, VERB=POST, uri=/upload/storage/v1/b/ not supported
	simpleDownload                     // VERB=GET, uri=/download
	streamingDownload                  // unsupported
	passThru                           // all other requests
)

func InterceptGcsMethod(f *proxy.Flow) gcsMethod {
	if f.Request.URL.Host == "storage.googleapis.com" {
		if strings.HasPrefix(f.Request.URL.Path, "/upload/storage/v1/b/") {
			if f.Request.Method == "POST" {
				if f.Request.URL.Query().Get("uploadType") == "multipart" {
					return multiPartUpload
				}
				if f.Request.URL.Query().Get("uploadType") == "media" {
					return singlePartUpload
				}
				if f.Request.URL.Query().Get("uploadType") == "resumable" {
					return resumableUpload
				}
			}
		}

		if strings.HasPrefix(f.Request.URL.Path, "/download") {
			if f.Request.Method == "GET" {
				return simpleDownload

			}
		}
	}
	return passThru
}

func (c *EncryptGcsPayload) Request(f *proxy.Flow) {

	var err error

out:
	switch m := InterceptGcsMethod(f); m {

	case multiPartUpload:
		// Parse the multipart request.
		err = HandleMultipartRequest(f)
		break out

	case simpleDownload:
		if f.Response != nil && f.Response.StatusCode == 404 {
			err = fmt.Errorf("404 detected '%v'", f.Request.Body)
		}
		//rangeReq := "bytes=0-" + strconv.Itoa(int(f.Request.Raw().ContentLength-1)) // breaks at file sizes bigger than 4GB
		//f.Request.Header.Set("range", rangeReq)
		break out

	case singlePartUpload:
		break out
	}

	if err != nil {
		log.Error(err)
		return
	}
}

func (c *DecryptGcsPayload) Response(f *proxy.Flow) {

	var err error

out:
	switch m := InterceptGcsMethod(f); m {

	case multiPartUpload:
		err = HandleMultipartResponse(f)
		break out

	case simpleDownload:
		err = HandleSimpleDownloadResponse(f)
		break out

	case singlePartUpload:
		break out
	}
	if err != nil {
		log.Error(err)
		return
	}
}
