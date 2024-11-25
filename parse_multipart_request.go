package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"

	"github.com/lqqyt2423/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

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

func ParseMultipartRequest(f *proxy.Flow) (*bytes.Buffer, *bytes.Buffer, error) {

	// Extract the boundary from the Content-Type header.
	contentType := f.Request.Header.Get("Content-Type")
	boundary := strings.Split(contentType, "boundary=")[1]
	boundary = strings.Trim(boundary, "'")

	// setup the body content reader
	bodyReader := strings.NewReader(string(f.Request.Body))

	multipartReader := multipart.NewReader(bodyReader, boundary)
	encryptedRequest := &bytes.Buffer{} //
	unencryptedFileContent := &bytes.Buffer{}

	// Creates a new multipart Writer with a random boundary, writing to the empty
	// buffer
	multipartWriter := multipart.NewWriter(encryptedRequest)

	err := multipartWriter.SetBoundary(boundary)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to set boundry in multipart-request: %v", err)
	}

	//Grab the first part. this contains the json metadata for the GCS request object
	part, err := multipartReader.NextPart()
	if err != nil {
		log.Fatal(err) //TODO change this to return error value
	}

	// Create the first part
	// grab the mime type for first part (should be application/json)
	// Process the part, get header , part value
	mimeHeader := GetMultipartMimeHeader(part)
	fmt.Println(mimeHeader)
	writer_part, err := multipartWriter.CreatePart(mimeHeader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new part in multipart-request: %v", err)
	}

	// Grab the actual JSON
	gcsObjectMetadataJson, err := io.ReadAll(part)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to json parse gcs object metadata: %v", err)
	}

	// unmarshall the json contents of the first part.
	var gcsObjectMetadataMap map[string]interface{}
	err = json.Unmarshal(gcsObjectMetadataJson, &gcsObjectMetadataMap)
	if err != nil {
		return nil, nil, fmt.Errorf("error unmarshalling gcsObjectMetadata: %v", err)
	}
	fmt.Println(gcsObjectMetadataMap)

	// store some extra metadata in GCS to help us on later requests
	//gcsObjectMetadataMap["x-unencrypted-content-length"] = string(len(f.Request.Body))
	//gcsObjectMetadataMap["x-md5Hash"] = ""
	//gcsObjectMetadataMap["x-tink-encryption"] = "1"

	// Now write the gcs object metadata back to the multipart writer
	jsonData, err := json.Marshal(gcsObjectMetadataMap)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshalling gcsObjectMetadata: %v", err)
	}
	writer_part.Write(jsonData)

	//Grab the second part. this contains the unencrypted file content
	part, err = multipartReader.NextPart()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading  multipart request: %v", err)
	}
	// Create the second part
	// the content-type here will always be  application/octet stream because we are storing encrypted
	// TODO ask eric about this, because we have to use the correct mime type or we get an error....
	///    writer_part, err = writer.CreatePart(GetMultipartMimeHeaderOctetStream())
	writer_part, err = multipartWriter.CreatePart(GetMultipartMimeHeader(part))
	if err != nil {
		return nil, nil, fmt.Errorf("error creating  multipart request: %v", err)
	}

	// Get file contents
	if part.FileName() == "" {
		rawBytes, err := io.ReadAll(part)
		unencryptedFileContent = bytes.NewBuffer(rawBytes)

		if err != nil {
			return nil, nil, fmt.Errorf("error reading  multipart request: %v", err)
		}

		// Encrypt the intercepted file
		encryptedData, err := encryptBytes(f.Request.Raw().Context(),
			config.KmsResourceName,
			unencryptedFileContent.Bytes())

		if err != nil {
			return nil, nil, fmt.Errorf("error encrypting  request: %v", err)
		}

		// write the final encrypted part
		writer_part.Write(encryptedData)
	}
	multipartWriter.Close()

	return encryptedRequest, unencryptedFileContent, nil

}
