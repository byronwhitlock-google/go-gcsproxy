package main

import (
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
	resumableUpload                    // uploadType=resumable, VERB=POST, uri=/upload/storage/v1/b/ not supported
	simpleDownload                     // VERB=GET, uri=/.... TODO
	streamingDownload                  // unsupported
	passThru                           // all other requests
)

func InterceptGcsMethod(f *proxy.Flow) gcsMethod {
	if f.Request.URL.Host == "storage.googleapis.com" &&
		strings.HasPrefix(f.Request.URL.Path, "/upload/storage/v1/b/") &&
		f.Request.Method == "POST" {
		if f.Request.URL.Query().Get("uploadType") == "multipart" {
			return multiPartUpload
		}
		if f.Request.URL.Query().Get("uploadType") == "media" {
			return singlePartUpload
		}
	}
	return passThru
}

func (c *EncryptGcsPayload) Request(f *proxy.Flow) {

	switch m := InterceptGcsMethod(f); m {
	case multiPartUpload:
		// Parse the multipart request.
		err := HandleMultipartRequest(f)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

func (c *DecryptGcsPayload) Response(f *proxy.Flow) {

	if InterceptGcsMethod(f) == multiPartUpload {
		// Parse the multipart request.
		err := HandleMultipartResponse(f)
		if err != nil {
			log.Error(err)
			return
		}
	}
}
