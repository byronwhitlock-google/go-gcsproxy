package main

import (
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
	multiPartUpload     gcsMethod = iota // uploadType=multipart, VERB=POST, path=/upload/storage/v1/b/  DOCS: https://cloud.google.com/storage/docs/json_api/v1/objects/insert
	singlePartUpload                     // uploadType=media,     VERB=POST, path=/upload/storage/v1/b/
	resumableUploadPost                  // uploadType=resumable, VERB=POST, path=/upload/storage/v1/b/
	resumableUploadPut                   // uploadType=resumable, VERB=PUT , path=/upload/storage/v1/b/
	simpleDownload                       // VERB=GET, path=/storage/v1/b/bucket/o/object?alt=media or path=/bucket-name/object-name
	streamingDownload                    // unsupported
	metadataRequest                      // VERB=GET, path=/storage/v1/b/bucket/o/object?alt=json or path=/storage/v1/b/bucket/o/object?fields=size,generation,updated
	passThru                             // all other requests

)

func IsEncryptDisabled() bool {
	if os.Getenv("GCS_PROXY_DISABLE_ENCRYPTION") == "" {
		return false
	}
	return true
}

func InterceptGcsMethod(f *proxy.Flow) gcsMethod {
	bucketName := getBucketNameFromRequestUri(f.Request.URL.Path)
	if getKMSKeyName(bucketName) == "" {
		return passThru
	}
	// GCS supports both hostnames
	if f.Request.URL.Host == "storage.googleapis.com" || f.Request.URL.Host == "www.googleapis.com" {
		// multi-part or simple upload
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

		// Resumable upload
		if strings.HasPrefix(f.Request.URL.Path, "/resumable/upload/storage/v1") ||
			(strings.HasPrefix(f.Request.URL.Path, "/upload/storage/v1") && f.Request.URL.Query().Get("uploadType") == "resumable") {
			if f.Request.Method == "POST" {
				return resumableUploadPost
			} else if f.Request.Method == "PUT" {
				return resumableUploadPut
			}
		}

		// get metadata
		if strings.HasPrefix(f.Request.URL.Path, "/storage/v1/b/") {
			if f.Request.Method == "GET" {
				// pass through for metadata request for bucket
				// TODO eshen may need to bypass directory too
				if strings.HasSuffix(f.Request.URL.Path, "/o") {
					return passThru
				}
				if f.Request.URL.Query().Get("alt") == "json" {
					return metadataRequest
				}
				if f.Request.URL.Query().Get("alt") == "media" {
					return simpleDownload
				}
				if f.Request.URL.Query().Get("fields") != "" {
					f.Request.URL.RawQuery = "alt=json"
					return metadataRequest
				}

			}
		}

		// download object when path=/download
		if strings.HasPrefix(f.Request.URL.Path, "/download") {
			return simpleDownload
		}
		// download when path=/bucket-name/object-name
		if f.Request.Method == "GET" {
			if f.Request.URL.Query().Get("alt") == "" || f.Request.URL.Query().Get("fields") == "" {
				return simpleDownload
			}

		}

	}
	return passThru
}

func (c *EncryptGcsPayload) Request(f *proxy.Flow) {

	debugRequest(f)
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

	debugResponse(f)

	if f.Response.StatusCode < 200 || f.Response.StatusCode > 299 {
		log.Errorf("got invalid response code! '%s' '%v'......\n\n%s", f.Request.URL, f.Response.StatusCode, f.Response.Body)
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

func debugResponse(f *proxy.Flow) {
	header := "<<<" + f.Id.String()
	log.Debugf("%v url: %v %v", header, f.Request.Method, f.Request.URL.String())
	log.Debugf("%v body len: %v, ", header, len(f.Response.Body))
	log.Debugf("%v header: %#v", header, f.Response.Header)
}

func debugRequest(f *proxy.Flow) {
	header := ">>>" + f.Id.String()
	log.Debugf("%v url: %v %v", header, f.Request.Method, f.Request.URL.String())
	log.Debugf("%v body len: %v, ", header, len(f.Request.Body))
	log.Debugf("%v header: %#v", header, f.Request.Header)
}
