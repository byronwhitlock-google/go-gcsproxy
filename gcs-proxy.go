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
	metadataRequest
	passThru // all other requests

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
			//if f.Request.Method == "GET" {
			return simpleDownload
			//}
		}

		if strings.HasPrefix(f.Request.URL.Path, "/storage/v1/b/") {
			if f.Request.Method == "GET" {
				if f.Request.URL.Query().Get("alt") == "json" {
					return metadataRequest
				}
			}
		}
	}
	return passThru
}

func (c *EncryptGcsPayload) Request(f *proxy.Flow) {

	log.Debug(fmt.Sprintf("got request: %s", f.Request.Raw().RequestURI))
	var err error

out:
	switch m := InterceptGcsMethod(f); m {

	case multiPartUpload:
		// Parse the multipart request.
		err = HandleMultipartRequest(f)
		break out

	case simpleDownload:
		HandleSimpleDownloadRequest(f)
		break out

	case singlePartUpload:
		break out

	case metadataRequest:
		HandleMetadataRequest(f)
		break out
	}
	if err != nil {
		log.Error(err)
		return
	}
}

func (c *DecryptGcsPayload) Response(f *proxy.Flow) {

	var err error

	if f.Response.StatusCode < 200 || f.Response.StatusCode > 299 {
		log.Error(fmt.Errorf("got invalid response code! '%v'......\n\n%s", f.Response.StatusCode, f.Response.Body))
	}
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

	case metadataRequest:
		HandleMetadataResponse(f)
		break out
	}
	if err != nil {
		log.Error(err)
		return
	}

	// recalculate content length
	f.Response.ReplaceToDecodedBody()
}
