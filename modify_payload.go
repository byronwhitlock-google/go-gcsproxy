package main

import (
	"fmt"
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

	//1 only MITM the storage.googleapis.com
	if f.Request.URL.Host != "storage.googleapis.com" {
		return
	}

	// only encrypt calls to the with GCS upload API
	if !strings.HasPrefix(f.Request.URL.Path, "/upload/storage/v1/b") {
		return
	}

	// ONLY look at post methods
	if f.Request.Method != "POST" {
		return
	}

	// we support

	fullBody := f.Request.Body

	println(fmt.Sprintf("This is the fullbody ! %s", fullBody))

	if strings.Contains(contentType, "text/html") {
		return
	}

	// change html <title> end with: " - go-mitmproxy"
	//f.Request.Raw().Response.Body
	//f.Response.Body = titleRegexp.ReplaceAll(f.Response.Body, []byte("${1}${2} - go-mitmproxy${3}"))
	//f.Response.Header.Set("Content-Length", strconv.Itoa(len(f.Response.Body)))
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
