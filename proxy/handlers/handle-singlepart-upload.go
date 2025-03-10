/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strconv"

	"github.com/byronwhitlock-google/go-gcsproxy/crypto"
	"github.com/byronwhitlock-google/go-gcsproxy/util"
	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

/*
	Steps to convert SinglePartUpload to MultiPartUpload:
		1. Change the url to use multipart in request url
		2. Change headers of request
		3. Change the body to use boundary and add metadata and body(ecnrypted)
*/

func ConvertSinglePartUploadtoMultiPartUpload(f *proxy.Flow) error {

	// URL change to use Multipart
	objectName := f.Request.URL.Query().Get("name")
	f.Request.URL.RawQuery = "uploadType=multipart&alt=json"

	//  Store original headers in variables, useful for generating metadata
	orgContentType := f.Request.Header.Get("Content-Type")

	f.Request.Method = "POST"
	log.Debugf("ConvertSinglePartUploadtoMultiPartUpload orgContentType: %v. Method changed to %v", orgContentType, f.Request.Method)

	//  Change headers to use multipart
	headersMap, boundary := util.GenerateHeadersList(f)
	for key, value := range headersMap {
		log.Debugf("%v: %v\n", key, value)
		f.Request.Header.Set(key, value)
	}

	f.Request.Header.Set("gcs-proxy-original-content-length",
		f.Request.Header.Get("Content-Length"))

	f.Request.Header.Set("gcs-proxy-unencrypted-file-size",
		strconv.Itoa(len(f.Request.Body)))

	// save the original md5 has or gsutil/gcloud will delete after upload if it sees it is different
	f.Request.Header.Set("gcs-proxy-original-md5-hash",
		crypto.Base64MD5Hash(f.Request.Body))

	f.Request.Header.Del("Expect")

	// Generate Metadata to insert in body
	metadata := util.GenerateMetadata(f, orgContentType, objectName)

	// Encrypt data in body
	bucketName := util.GetBucketNameFromRequestUri(f.Request.URL.Path)
	ctx := f.Request.Raw().Context()
	ctxValue := context.WithValue(ctx, "requestid", f.Id.String())
	encryptBody, err := crypto.EncryptBytes(ctxValue,
		util.GetKMSKeyName(bucketName),
		f.Request.Body)
	if err != nil {
		return fmt.Errorf("error encrypting  request: %v", err)
	}

	//Write data to request body  to support multipart request
	encryptedRequest := &bytes.Buffer{}
	multipartWriter := multipart.NewWriter(encryptedRequest)
	err = multipartWriter.SetBoundary(boundary)
	if err != nil {
		return fmt.Errorf("failed to set boundary in multipart-request: %v", err)
	}

	// Adding First part
	writer_part, err := multipartWriter.CreatePart(util.CreateFirstMultipartMimeHeader())
	if err != nil {
		return fmt.Errorf("failed to create first part in multipart-request: %v", err)
	}
	marshalled_metadata, err := json.Marshal(metadata)
	writer_part.Write(marshalled_metadata)

	// Adding second part
	writer_part, err = multipartWriter.CreatePart(util.CreateSecondMultipartMimeHeader(orgContentType))
	if err != nil {
		return fmt.Errorf("failed to create second part in multipart-request: %v", err)
	}
	writer_part.Write(encryptBody)

	multipartWriter.Close()

	// update the body to the newly encrypted request
	f.Request.Body = encryptedRequest.Bytes()

	return nil
}

func HandleSinglePartUploadResponse(f *proxy.Flow) error {
	var jsonResponse map[string]interface{}
	// turn the response body into a dynamic json map we can use
	err := json.Unmarshal(f.Response.Body, &jsonResponse)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	// update the response with the orginal md5 hash so gsutil/gcloud does not complain
	jsonResponse["md5Hash"] = f.Request.Header.Get("Gcs-proxy-original-md5-hash")
	jsonResponse["size"], err = strconv.Atoi(f.Request.Header.Get("Gcs-proxy-unencrypted-file-size"))
	if err != nil {
		return fmt.Errorf("error setting json response: %v", err)
	}

	log.Debugf("HandleSinglePartUploadResponse response with original size and md5: %v", jsonResponse)
	jsonData, err := json.Marshal(jsonResponse)
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	f.Response.Body = jsonData
	return nil
}

func HandleSinglePartUploadRequest(f *proxy.Flow) error {
	bucketName := util.GetBucketNameFromRequestUri(f.Request.URL.Path)
	ctx := f.Request.Raw().Context()
	ctxValue := context.WithValue(ctx, "requestid", f.Id.String())
	encryptedData, err := crypto.EncryptBytes(ctxValue,
		util.GetKMSKeyName(bucketName),
		f.Request.Body)

	if err != nil {
		return fmt.Errorf("error encrypting  request: %v", err)
	}

	f.Request.Header.Set("gcs-proxy-original-content-length",
		f.Request.Header.Get("Content-Length"))

	f.Request.Header.Set("Content-Length",
		strconv.Itoa(len(encryptedData)))

	// save the original md5 has or gsutil/gcloud will delete after upload if it sees it is different
	f.Request.Header.Set("gcs-proxy-original-md5-hash",
		crypto.Base64MD5Hash(f.Request.Body))

	f.Request.Header.Set("gcs-proxy-unencrypted-file-size",
		strconv.Itoa(len(f.Request.Body)))

	f.Request.Body = encryptedData

	return nil
}
