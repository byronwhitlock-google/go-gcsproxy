package main

import (
	"encoding/json"
	"fmt"
	"strconv"
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
type GCS_METHOD int

const (
	multiPartUpload   GCS_METHOD = iota // uploadType=multipart, VERB=POST, uri=/upload/storage/v1/b/  DOCS: https://cloud.google.com/storage/docs/json_api/v1/objects/insert
	singlePartUpload                    // uploadType=media,     VERB=POST, uri=/upload/storage/v1/b/
	resumableUpload                     // uploadType=resumable, VERB=POST, uri=/upload/storage/v1/b/ not supported
	simpleDownload                      // VERB=GET, uri=/.... TODO
	streamingDownload                   // unsupported
	passThru                            // all other requests
)

func InterceptGcsMethod(f *proxy.Flow) GCS_METHOD {
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

	if f.Request.URL.Host == "storage.googleapis.com" && 
		strings.HasPrefix(f.Request.URL.Path, "/download") &&
		f.Request.Method == "GET" {
			return simpleDownload
		}

	return passThru
}

func (c *EncryptGcsPayload) Request(f *proxy.Flow) {

	if InterceptGcsMethod(f) == multiPartUpload {

		// Parse the multipart request.
		// TODO Fix this mess of string parsing and use the native stream
		// TODO untangle parse multipart request from CreateMultipart request. We need to do this so we can unconditionally rewrite single part uploads to multipart in order to add extra gcs object metadata
		encrypted_request, unencrypted_file_content, err := ParseMultipartRequest(f)
		if err != nil {
			panic(err)
		}
		fmt.Println(unencrypted_file_content)
		fmt.Println("Hash Value at Request")
		fmt.Println(base64_md5hash(unencrypted_file_content.Bytes()))

		f.Request.Header.Set("gcs-proxy-original-content-length",
			string(len(f.Request.Body)))

		f.Request.Body = encrypted_request.Bytes()

		f.Request.Header.Set("gcs-proxy-original-md5-hash",
			base64_md5hash(unencrypted_file_content.Bytes()))
	}
}

func (c *DecryptGcsPayload) Response(f *proxy.Flow) {

	if InterceptGcsMethod(f) == multiPartUpload {
		fmt.Println("Multipart")

		var jsonResponse map[string]string
		// turn the response body into a dynamic json map we can use
		err := json.Unmarshal(f.Response.Body, &jsonResponse)
		if err != nil {
			log.Fatalf("Error unmarshalling JSON: %v", err)
		}
		fmt.Println(jsonResponse)

		// update the response with the orginal md5 hash so gsutil/gcloud does not complain
		jsonResponse["md5Hash"] = f.Request.Header.Get("gcs-proxy-original-md5-hash")

		jsonData, err := json.Marshal(jsonResponse)
		if err != nil {
			fmt.Println("Error marshaling to JSON:", err)
		}

		f.Response.Body = jsonData

		// recalculate content length
		f.Response.ReplaceToDecodedBody()
	}

	if InterceptGcsMethod(f) == simpleDownload {
		fmt.Println("simpleDownload")
		//fmt.Println(string(f.Response.Body))

		// Update the response content with the decrypted content
		original_content, err := decrypt_tink(f.Response.Body)
		if err != nil {
			fmt.Println("Unable to decrypt response body:", err)
			log.Fatal(err)
		}
		
		fmt.Println(original_content)
		fmt.Println(len(original_content))
		f.Response.Body = original_content
		content_length := len(f.Response.Body)
		content_length_str := strconv.Itoa(len(f.Response.Body))

		// Update content length headers with new length of decrypted data
		f.Response.Header.Set("X-Goog-Stored-Content-Length",
			content_length_str)
		
		f.Response.Header.Set("Content-Length",
			content_length_str)

		// gcloud storage cp command uses "range" in request
        range_value := "bytes 0-"+strconv.Itoa(content_length-1)+"/"+content_length_str

		f.Request.Header.Set("range",
			range_value)

		f.Response.Header.Set("Content-Range",
			range_value)
	   
		hash_value := base64_md5hash(original_content)
		f.Response.Header.Set("X-Goog-Hash",
			hash_value)

	}
}
