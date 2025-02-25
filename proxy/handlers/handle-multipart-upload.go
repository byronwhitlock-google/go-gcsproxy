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
	"io"
	"mime"
	"mime/multipart"
	"net/textproto"
	"strconv"
	"strings"

	cfg "github.com/byronwhitlock-google/go-gcsproxy/config"
	"github.com/byronwhitlock-google/go-gcsproxy/crypto"
	"github.com/byronwhitlock-google/go-gcsproxy/util"
	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

// ##### Not using ######
func GetMultipartMimeHeaderOctetStream() textproto.MIMEHeader {
	// Process the part, get header , part value
	mimeHeader := textproto.MIMEHeader{}
	mimeHeader.Set("Content-Type", "application/octet-stream")

	return mimeHeader
}

func GetMultipartMimeHeader(part *multipart.Part) textproto.MIMEHeader {
	// Process the part, get header , part value
	mimeHeader := textproto.MIMEHeader{}
	//Loop through Map
	for k, v := range part.Header {
		mimeHeader.Set(k, v[0])
	}
	return mimeHeader
}

func HandleMultipartRequest(f *proxy.Flow) error {

	// Extract the boundary from the Content-Type header.
	contentType := f.Request.Header.Get("Content-Type")
	log.Debugf("in HandleMultipartRequest, got content-type: %v", contentType)

	// Remove single quotes from the boundary parameter
	// RFC 2046 (MIME) is the key document for multipart messages. The boundary is a RFC822 parameter
	//RFC 822 defines parameters as "attribute = value" where value can be a token or a quoted-string.
	//RFC 822 only defines quoted-string with double quotes, not single quotes.
	contentType = strings.ReplaceAll(contentType, "'", "\"")

	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("error parsing content type %v", err)
	}
	boundary := params["boundary"]

	// setup the body content reader
	bodyReader := strings.NewReader(string(f.Request.Body))

	multipartReader := multipart.NewReader(bodyReader, boundary)
	encryptedRequest := &bytes.Buffer{} //
	unencryptedFileContent := &bytes.Buffer{}

	// Creates a new multipart Writer with a random boundary, writing to the empty
	// buffer
	multipartWriter := multipart.NewWriter(encryptedRequest)

	err = multipartWriter.SetBoundary(boundary)
	if err != nil {
		return fmt.Errorf("failed to set boundary in multipart-request: %v", err)
	}

	//Grab the first part. this contains the json metadata for the GCS request object
	part, err := multipartReader.NextPart()
	if err != nil {
		return fmt.Errorf("failed to read next part in multipart-request: %v", err)
	}

	// Create the first part
	// grab the mime type for first part (should be application/json)
	// Process the part, get header , part value
	mimeHeader := GetMultipartMimeHeader(part)
	log.Debug(mimeHeader)
	writer_part, err := multipartWriter.CreatePart(mimeHeader)
	if err != nil {
		return fmt.Errorf("failed to create new part in multipart-request: %v", err)
	}

	// Grab the actual JSON
	gcsObjectMetadataJson, err := io.ReadAll(part)
	if err != nil {
		return fmt.Errorf("failed to json parse gcs object metadata: %v", err)
	}

	// TODO: pull in the gcs sdk so we have an up to date proto
	var gcsMetadata interface{}
	// unmarshall the json contents of the first part.
	err = json.Unmarshal(gcsObjectMetadataJson, &gcsMetadata)
	if err != nil {
		return fmt.Errorf("error unmarshalling gcsObjectMetadata: %v", err)
	}

	gcsMetadataMap, ok := gcsMetadata.(map[string]interface{})
	if !ok {
		return fmt.Errorf("error: JSON data is not a map")
	}
	if gcsMetadataMap["metadata"] == nil {
		gcsMetadataMap["metadata"] = make(map[string]interface{})
	}

	bucketName := util.GetBucketNameFromGcsMetadata(gcsMetadataMap)
	if bucketName == "" {
		bucketName = util.GetBucketNameFromRequestUri(f.Request.URL.Path)
	}

	//Grab the second part. this contains the unencrypted file content
	part, err = multipartReader.NextPart()
	if err != nil {
		return fmt.Errorf("error reading  multipart request: %v", err)
	}

	var encryptedData []byte
	// Get file contents
	if part.FileName() == "" {
		rawBytes, err := io.ReadAll(part)
		unencryptedFileContent = bytes.NewBuffer(rawBytes)

		if err != nil {
			return fmt.Errorf("error reading  multipart request: %v", err)
		}

		// Encrypt the intercepted file

		ctx := f.Request.Raw().Context()
		ctxValue := context.WithValue(ctx, "requestid", f.Id.String())
		encryptedData, err = crypto.EncryptBytes(ctxValue,
			util.GetKMSKeyName(bucketName),
			unencryptedFileContent.Bytes())

		if err != nil {
			return fmt.Errorf("error encrypting  request: %v", err)
		}

	}
	///
	///
	/// Create multipart request
	///
	///
	// TODO move this into its own method
	// Access and modify the nested value dynamically
	customMetadata, ok := gcsMetadataMap["metadata"].(map[string]interface{})
	if ok {

		customMetadata["x-unencrypted-content-length"] = len(unencryptedFileContent.String())
		customMetadata["x-md5Hash"] = crypto.Base64MD5Hash(unencryptedFileContent.Bytes())
		customMetadata["x-encryption-key"] = util.GetKMSKeyName(bucketName)
		customMetadata["x-proxy-version"] = cfg.GlobalConfig.GCSProxyVersion // TODO: Change this to the global Version in the main package ASAP
	}

	log.Debug(string(gcsObjectMetadataJson))
	log.Debug(gcsMetadata)
	log.Debug(fmt.Errorf("got metadata: %s", gcsObjectMetadataJson))

	// Now write the gcs object metadata back to the multipart writer
	newGcsMetadataJson, err := json.Marshal(gcsMetadata)

	if err != nil {
		return fmt.Errorf("error marshalling gcsObjectMetadata: %v", err)
	}
	log.Debug(fmt.Errorf("rewrote json data to: %s", newGcsMetadataJson))

	writer_part.Write(newGcsMetadataJson)

	// Create the second part
	// the content-type here will always be  application/octet stream because we are storing encrypted
	// TODO ask eric about this, because we have to use the correct mime type or we get an error....
	///    writer_part, err = writer.CreatePart(GetMultipartMimeHeaderOctetStream())
	writer_part, err = multipartWriter.CreatePart(GetMultipartMimeHeader(part))
	if err != nil {
		return fmt.Errorf("error creating  multipart request: %v", err)
	}

	// write the final encrypted part
	writer_part.Write(encryptedData)

	multipartWriter.Close()

	// Save the original content length for rewriting when download.
	f.Request.Header.Set("gcs-proxy-original-content-length",
		f.Request.Header.Get("Content-Length"))

	f.Request.Header.Set("gcs-proxy-unencrypted-file-size",
		strconv.Itoa(unencryptedFileContent.Len()))

	log.Trace(unencryptedFileContent)
	log.Trace(encryptedRequest)

	// update the body to the newly encrypted request
	f.Request.Body = encryptedRequest.Bytes()

	// save the original md5 has or gsutil/gcloud will delete after upload if it sees it is different
	f.Request.Header.Set("gcs-proxy-original-md5-hash",
		crypto.Base64MD5Hash(unencryptedFileContent.Bytes()))

	return nil
}

func HandleMultipartResponse(f *proxy.Flow) error {
	log.Debug("in HandleMultipartResponse")

	var jsonResponse map[string]interface{}
	// turn the response body into a dynamic json map we can use
	err := json.Unmarshal(f.Response.Body, &jsonResponse)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	log.Debug(jsonResponse)

	// update the response with the orginal md5 hash so gsutil/gcloud does not complain
	jsonResponse["md5Hash"] = f.Request.Header.Get("gcs-proxy-original-md5-hash")
	jsonResponse["size"], err = strconv.Atoi(f.Request.Header.Get("gcs-proxy-unencrypted-file-size"))
	if err != nil {
		return fmt.Errorf("error setting json response: %v", err)
	}

	jsonData, err := json.Marshal(jsonResponse)
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	f.Response.Body = jsonData
	return nil
}
