package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

type DecryptGcsPayload struct {
	proxy.BaseAddon
}
type EncryptGcsPayload struct {
	proxy.BaseAddon
}

type GetReqHeader struct {
	proxy.BaseAddon
}

// https://cloud.google.com/storage/docs/json_api/v1/objects
type gcsMethod int

const (
	multiPartUpload     gcsMethod = iota // uploadType=multipart, VERB=POST, uri=/upload/storage/v1/b/  DOCS: https://cloud.google.com/storage/docs/json_api/v1/objects/insert
	singlePartUpload                     // uploadType=media,     VERB=POST, uri=/upload/storage/v1/b/
	resumableUploadPost                  // unsupported uploadType=resumable, VERB=POST, uri=/upload/storage/v1/b/
	resumableUploadPut                   // unsupported uploadType=resumable, VERB=PUT , uri=/upload/storage/v1/b/
	simpleDownload                       // VERB=GET, uri=/download
	streamingDownload                    // unsupported
	metadataRequest
	passThru // all other requests

)

func IsEncryptDisabled() bool {
	if os.Getenv("GCS_PROXY_DISABLE_ENCRYPTION") == "" {
		return false
	}
	return true
}

func InterceptGcsMethod(f *proxy.Flow) gcsMethod {
	if f.Request.URL.Host == "storage.googleapis.com" {
		if strings.HasPrefix(f.Request.URL.Path, "/upload/storage/v1") {
			if f.Request.Method == "POST" {
				if f.Request.URL.Query().Get("uploadType") == "multipart" {
					return multiPartUpload
				}
				if f.Request.URL.Query().Get("uploadType") == "media" {
					return singlePartUpload
				}
			}
		}
		if strings.HasPrefix(f.Request.URL.Path, "/resumable/upload/storage/v1") || strings.HasPrefix(f.Request.URL.Path, "/upload/storage/v1") {
			if f.Request.Method == "POST" {
				return resumableUploadPost
			} else if f.Request.Method == "PUT" {
				return resumableUploadPut
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

func (h *GetReqHeader) Requestheaders(f *proxy.Flow) {
	log.Debug(fmt.Sprintf("got request headers: %s", f.Request.Raw().Header))
}

func (c *EncryptGcsPayload) Request(f *proxy.Flow) {

	log.Debug(fmt.Sprintf("got request: %s", f.Request.Raw().RequestURI))
	if IsEncryptDisabled() {
		return
	}

	var err error

out:
	switch m := InterceptGcsMethod(f); m {

	case multiPartUpload:
		// Parse the multipart request.
		err = HandleMultipartRequest(f)
		break out

	case simpleDownload:
		err = HandleSimpleDownloadRequest(f)
		break out

	case singlePartUpload:
		err = ConvertSinglePartUploadtoMultiPartUpload(f)
		break out

	case metadataRequest:
		err = HandleMetadataRequest(f)
		break out

	case resumableUploadPost:
		err = HandleResumablePostRequest(f)
		break out

	case resumableUploadPut:
		err = HandleResumablePutRequest(f)
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
		log.Error(fmt.Errorf("got invalid response code! '%s' '%v'......\n\n%s", f.Request.URL, f.Response.StatusCode, f.Response.Body))
	}
	if IsEncryptDisabled() {
		return
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
		err = HandleSinglePartUploadResponse(f)
		break out

	case metadataRequest:
		err = HandleMetadataResponse(f)
		break out

	case resumableUploadPost:
		err = HandleResumablePostResponse(f)
		break out

	case resumableUploadPut:
		err = HandleResumablePutResponse(f)
		break out

	}
	if err != nil {
		log.Error(err)
		return
	}

	// recalculate content length
	f.Response.ReplaceToDecodedBody()
}

func (c *EncryptGcsPayload) StreamRequestModifier(f *proxy.Flow, io io.Reader) io.Reader{
	fmt.Println("In StreamRequestModifier")
	fmt.Println(f)
	fmt.Println(io)
	stringReader := strings.NewReader("Maximum object size reached. Stream Processing in proxy disabled. Proxy supports files for a maximum size of 64GB. ")
	return stringReader 
}

func (c *DecryptGcsPayload) StreamResponseModifier(f *proxy.Flow, io io.Reader) io.Reader{
	fmt.Println("In StreamResponseModifier")
	fmt.Println(f)
	fmt.Println(io)
	stringReader := strings.NewReader("Maximum object size reached. Stream Processing in proxy disabled. Proxy supports files for a maximum size of 64GB. ")
	return stringReader 
	
}
