package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lqqyt2423/go-mitmproxy/proxy"
)

type DecryptGcsPayload struct {
	proxy.BaseAddon
}
type EncryptGcsPayload struct {
	proxy.BaseAddon
}

func (c *EncryptGcsPayload) Request(f *proxy.Flow) {

	contentType := f.Request.Header.Get("Content-Type")
	// https://cloud.google.com/storage/docs/json_api/v1/objects

	// We are handling insert
	// https://cloud.google.com/storage/docs/json_api/v1/objects/insert
	/*
		POST https://storage.googleapis.com/upload/storage/v1/b/bucket/o
	*/

	//1 only MITM the storage.googleapis.com
	if f.Request.URL.Host != "storage.googleapis.com" {
		return
	}
	// only encrypt calls to the with GCS upload API
	if !strings.HasPrefix(f.Request.URL.Path, "/upload/storage/v1/b/") {
		return
	}

	//ONLY look at post methods
	// NOTE: PUT methods are for resumable downloads
	if f.Request.Method != "POST" {
		return
	}

	// we support uploadType=multipart
	qs := f.Request.URL.Query()
	if qs.Get("uploadType") == "multipart" {

		// Extract the boundary from the Content-Type header.
		boundary := strings.Split(contentType, "boundary=")[1]
		boundary = strings.Trim(boundary, "'")

		// Parse the multipart request.
		// TODO Fix this mess of string parsing and use the native stream
		gcs_metadata, fileContent, err := ParseMultipartRequest(strings.NewReader(string(f.Request.Body)), boundary)
		if err != nil {
			panic(err)
		}

		//fmt.Println("Metadata:", gcs_metadata)
		//fmt.Println("File Content:", fileContent)
		//fmt.Println("Encrytion started")
		ciphertext , err:= encrypt_tink(fileContent)
		if err != nil {
			panic(err)
		}
		fmt.Println("Encrytion done")
		//fmt.Println(f.Request.Raw().Response.Body)
		//fmt.Println(f.Response)

		f.Request.Header.Set("Content-Type", "application/octet-stream")
		//encryptedBody := []byte(ciphertext) // Replace with your encryption function

			// Set the new body as an io.ReadCloser
		//reader := strings.NewReader(ciphertext)
		fullbody:= string(gcs_metadata)+"\n"+ciphertext
		f.Request.Body = []byte(fullbody)  //Unbale to replace
		//f.Request.Body = []byte(ciphertext) //titleRegexp.ReplaceAll(string(ciphertext), []byte("${1}${2} - go-mitmproxy${3}"))
		//f.Request.Body = titleRegexp.ReplaceAll(ciphertext, []byte("${1}${2} - go-mitmproxy${3}"))
	}

	//fullBody := f.Request.Body

	//println(fmt.Sprintf("This is the fullbody ! %s", fullBody))

	if strings.Contains(contentType, "text/html") {
		return
	}

	// change html <title> end with: " - go-mitmproxy"
	//f.Request.Raw().Response.Body
	//f.Response.Body = titleRegexp.ReplaceAll(f.Response.Body, []byte("${1}${2} - go-mitmproxy${3}"))
	f.Request.Header.Set("Content-Length", strconv.Itoa(len(f.Request.Body)))
	fmt.Println(f.Request.Body)
}

func (c *DecryptGcsPayload) Response(f *proxy.Flow) {
	contentType := f.Response.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		return
	}

	// change html <title> end with: " - go-mitmproxy"
	f.Response.ReplaceToDecodedBody()
	//f.Response.Body = titleRegexp.ReplaceAll(f.Response.Body, []byte("${1}${2} - go-mitmproxy${3}"))
	//f.Response.Header.Set("Content-Length", strconv.Itoa(len(f.Response.Body)))
}
