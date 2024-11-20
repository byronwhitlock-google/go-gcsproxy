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

var boundary string
//var original_content []byte
// var body *bytes.Buffer
// var err error
var org_encoded_str string
var checksum uint32


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
		boundary = strings.Split(contentType, "boundary=")[1]
		boundary = strings.Trim(boundary, "'")

		// Parse the multipart request.
		// TODO Fix this mess of string parsing and use the native stream
		body,original_content,err := ParseMultipartRequest(strings.NewReader(string(f.Request.Body)), boundary)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(original_content))
		fmt.Println(body)
		f.Request.Header.Set("gcs-proxy-original-content-length",string(len(f.Request.Body)))
		//f.Request.Header.Set("Content-Type", "application/octet-stream")
		//f.Request.Body = body.Bytes()
		org_encoded_str=base64_md5hash(original_content)

		f.Request.Header.Set("gcs-proxy-original-md5-hash",org_encoded_str)

	}

	if strings.Contains(contentType, "text/html") {
		return
	}
	
	f.Request.Header.Set("Content-Length", strconv.Itoa(len(f.Request.Body)))
	
}

func (c *DecryptGcsPayload) Response(f *proxy.Flow) {
	contentType := f.Response.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		return
	}

	// change html <title> end with: " - go-mitmproxy"
	//f.Response.ReplaceToDecodedBody()

	//contentType_req := f.Request.Header.Get("Content-Type")

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
		fmt.Println("Multipart")
		
		
		var result map[string]string
		err:=json.Unmarshal(f.Response.Body, &result)
		if err != nil {
			log.Fatalf("Error unmarshalling JSON: %v", err)
		}
		fmt.Println(result)

		result["md5Hash"]=org_encoded_str
		//result["size"]=string(len(org_encoded_str))
		


		jsonData, err := json.Marshal(result)
			if err != nil {
						fmt.Println("Error marshaling to JSON:", err)
			}

		//fmt.Println(boundary)
		//f.Request.Header.Set("Content-Type", "application/octet-stream")
		
		// org_encoded_str:=base64_md5hash(original_content)
		f.Response.Body = jsonData//bytes.NewBuffer(result)
		//fmt.Println(string(f.Response.Body))
	}
	//fmt.Println(f.Response.Body)


	//f.Response.Body = titleRegexp.ReplaceAll(f.Response.Body, []byte("${1}${2} - go-mitmproxy${3}"))
	//f.Response.Header.Set("Content-Length", strconv.Itoa(len(f.Response.Body)))
}

//"{\n  \"generation\": \"1732123587119115\",\n  \"size\": \"90\",\n  \"md5Hash\": \"dXyGhoMrnN/IxJ/EZy2hpg==\",\n  \"crc32c\": \"BQ/ZuQ==\",\n  \"etag\": \"CIuQ+Ji364kDEAE=\"\n}\n"

//"{\"crc32c\":\"BQ/ZuQ==\",\"etag\":\"CIuQ+Ji364kDEAE=\",\"generation\":\"1732123587119115\",\"md5Hash\":\"XgtItMwnTfdnkKaJ6bIFOA==\",\"size\":\"\\u0018\"}"
//"{\"crc32c\":\"q0tWpA==\",\"etag\":\"CPj1pNW064kDEAE=\",\"generation\":\"1732122908375800\",\"md5Hash\":\"XgtItMwnTfdnkKaJ6bIFOA==\",\"size\":\"90\"}"


