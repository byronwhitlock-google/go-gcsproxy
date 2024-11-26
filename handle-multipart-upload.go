package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
	"strconv"

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

func HandleMultipartRequest(f *proxy.Flow) error {

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
		return fmt.Errorf("failed to set boundry in multipart-request: %v", err)
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
		return fmt.Errorf("failed to create new part in multipart-request: %v", err)
	}

	// Grab the actual JSON
	gcsObjectMetadataJson, err := io.ReadAll(part)
	if err != nil {
		return fmt.Errorf("failed to json parse gcs object metadata: %v", err)
	}

	// unmarshall the json contents of the first part.
	var gcsObjectMetadataMap map[string]interface{}
	err = json.Unmarshal(gcsObjectMetadataJson, &gcsObjectMetadataMap)
	if err != nil {
		return fmt.Errorf("error unmarshalling gcsObjectMetadata: %v", err)
	}
	fmt.Println(gcsObjectMetadataMap)

	// store some extra metadata in GCS to help us on later requests
	//gcsObjectMetadataMap["x-unencrypted-content-length"] = string(len(f.Request.Body))
	//gcsObjectMetadataMap["x-md5Hash"] = ""
	//gcsObjectMetadataMap["x-tink-encryption"] = "1"

	// Now write the gcs object metadata back to the multipart writer
	jsonData, err := json.Marshal(gcsObjectMetadataMap)
	if err != nil {
		return fmt.Errorf("error marshalling gcsObjectMetadata: %v", err)
	}
	writer_part.Write(jsonData)

	//Grab the second part. this contains the unencrypted file content
	part, err = multipartReader.NextPart()
	if err != nil {
		return fmt.Errorf("error reading  multipart request: %v", err)
	}
	// Create the second part
	// the content-type here will always be  application/octet stream because we are storing encrypted
	// TODO ask eric about this, because we have to use the correct mime type or we get an error....
	///    writer_part, err = writer.CreatePart(GetMultipartMimeHeaderOctetStream())
	writer_part, err = multipartWriter.CreatePart(GetMultipartMimeHeader(part))
	if err != nil {
		return fmt.Errorf("error creating  multipart request: %v", err)
	}

	// Get file contents
	if part.FileName() == "" {
		rawBytes, err := io.ReadAll(part)
		unencryptedFileContent = bytes.NewBuffer(rawBytes)

		if err != nil {
			return fmt.Errorf("error reading  multipart request: %v", err)
		}

		// Encrypt the intercepted file
		encryptedData, err := encryptBytes(f.Request.Raw().Context(),
			config.KmsResourceName,
			unencryptedFileContent.Bytes())

		if err != nil {
			return fmt.Errorf("error encrypting  request: %v", err)
		}

		// write the final encrypted part
		writer_part.Write(encryptedData)
	}
	multipartWriter.Close()

	// Save the orginal content length for rewriting when download.
	f.Request.Header.Set("gcs-proxy-original-content-length",
		string(len(f.Request.Body)))

	log.Debug(unencryptedFileContent)
	log.Debug(encryptedRequest)

	// update the body to the newly encrypted request
	f.Request.Body = encryptedRequest.Bytes()

	// save the original md5 has or gsutil/gcloud will delete after upload if it sees it is different
	f.Request.Header.Set("gcs-proxy-original-md5-hash",
		base64_md5hash(unencryptedFileContent.Bytes()))

	return nil
}

func HandleMultipartResponse(f *proxy.Flow) error {
	log.Debug("in HandleMultipartResponse")

	var jsonResponse map[string]string
	// turn the response body into a dynamic json map we can use
	err := json.Unmarshal(f.Response.Body, &jsonResponse)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	fmt.Println(jsonResponse)

	// update the response with the orginal md5 hash so gsutil/gcloud does not complain
	jsonResponse["md5Hash"] = f.Request.Header.Get("gcs-proxy-original-md5-hash")

	jsonData, err := json.Marshal(jsonResponse)
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	//fmt.Println(jsonData)
	f.Response.Body = jsonData

	// recalculate content length
	f.Response.ReplaceToDecodedBody()
	return nil
}

func HandleSimpleDownloadResponse(f *proxy.Flow) error {
		fmt.Println("simpleDownload")

		// Update the response content with the decrypted content
		original_content, err := decryptBytes(f.Request.Raw().Context(),
			config.KmsResourceName,f.Response.Body)
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
        range_value_resp := "bytes 0-"+strconv.Itoa(content_length-1)+"/"+content_length_str
		range_value_req:= "bytes=0-"+strconv.Itoa(content_length-1)

		f.Request.Header.Set("range",
			range_value_req)

		f.Response.Header.Set("Content-Range",
			range_value_resp)
	   
		hash_value := base64_md5hash(original_content)
		f.Response.Header.Set("X-Goog-Hash",
			hash_value)
		
		return nil

}
