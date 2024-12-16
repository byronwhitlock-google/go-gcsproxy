package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"

	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

func HandleResumablePostRequest(f *proxy.Flow) error {
	// strip X-upload-content-length
	f.Request.Header.Del("x-upload-content-length")
	f.Request.Header.Del("X-Upload-Content-Length")
	return nil //do nothing
}

// this is the raw data to be encoded.
func HandleResumablePutRequest(f *proxy.Flow) error {

	// first we need the uploader id so we can get the resumable metadata.
	log.Debug(fmt.Sprintf("got query string  %s", f.Request.URL.RawQuery))

	uploadId := f.Request.URL.Query().Get("upload_id")

	if uploadId == "" {
		return fmt.Errorf(fmt.Sprintf("missing upload id in query string: %s", f.Request.URL.RawQuery))
	}
	resumeData, err := LoadResumableData(uploadId)

	if err != nil {
		return fmt.Errorf("error Loading Resumable Data: %v", err)
	}

	byteRangeHeader := f.Request.Header.Get("Content-Range")
	start, end, size, err := parseByteRangeHeader(byteRangeHeader)
	if err != nil {
		return err
	}

	if !(start == 0 && end+1 == size) {
		return fmt.Errorf("unsupported Byte range detected '%v'", byteRangeHeader)
	}

	ConvertSinglePartUploadtoMultiPartUpload(f, resumeData["name"])

	return nil
}

func HandleResumablePutResponse(f *proxy.Flow) error {
	log.Debug("in HandleResumablePutResponse")

	var jsonResponse map[string]interface{}
	// turn the response body into a dynamic json map we can use
	err := json.Unmarshal(f.Response.Body, &jsonResponse)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	fmt.Println(jsonResponse)

	// update the response with the original md5 hash so gsutil/gcloud does not complain
	jsonResponse["md5Hash"] = f.Request.Header.Get("gcs-proxy-original-md5-hash")
	jsonResponse["size"], err = strconv.Atoi(f.Request.Header.Get("gcs-proxy-unencrypted-file-size"))
	if err != nil {
		return fmt.Errorf("error setting json response: %v", err)
	}

	jsonData, err := json.Marshal(jsonResponse)
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %v", err)
	}

	//fmt.Println(jsonData)
	f.Response.Body = jsonData
	return nil
}

// rangeString = "bytes 0-72355493/72355494"
func parseByteRangeHeader(rangeStr string) (start int, end int, size int, err error) {
	// Regular expression to capture the start, end, and total values
	re := regexp.MustCompile(`bytes (\d+)-(\d+)/(\d+)`)
	matches := re.FindStringSubmatch(rangeStr)

	if len(matches) != 4 {
		return 0, 0, 0, fmt.Errorf("invalid range format: %s", rangeStr)
	}

	rStart, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error parsing start: %v", err)
	}

	rEnd, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error parsing end: %v", err)
	}

	rTotal, err := strconv.Atoi(matches[3])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("error parsing total: %v", err)
	}

	return rStart, rEnd, rTotal, nil
}

func HandleResumablePostResponse(f *proxy.Flow) error {

	log.Debug(fmt.Sprintf("Got metadata request: %s", f.Request.Body))

	// Unmarshal the json contents of the first part.
	var dataMap map[string]interface{}
	err := json.Unmarshal(f.Response.Body, &dataMap)
	if err != nil {
		return fmt.Errorf("error unmarshalling gcsObjectMetadata: %v", err)
	}

	uploaderId := f.Response.Header.Get("X-GUploader-UploadID")
	if uploaderId == "" {
		return fmt.Errorf("missing X-GUploader-UploadID header")
	}

	StoreResumableData(uploaderId, dataMap)

	return nil
}

// writes data to a file by id
func StoreResumableData(id string, dataMap map[string]interface{}) error {

	// use /tmp
	filePath := fmt.Sprint("/tmp/go-gcsproxy-%v.json", id)

	// Open the file for writing (creates the file if it doesn't exist)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file in StoreResumableData: %v", err)
	}
	defer file.Close() // Ensure the file is closed when the function exits

	// Now write the gcs object metadata back to the multipart writer
	jsonData, err := json.Marshal(dataMap)
	if err != nil {
		return fmt.Errorf("error marshalling ResumableData: %v", err)
	}

	// Write a string to the file
	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error writing file in StoreResumableData: %v", err)
	}

	// Flush any buffered data to the file
	file.Sync()

	log.Debug(fmt.Sprintf("wrote ResumableData: %s", jsonData))
	return nil
}

// reads data from a file by id
func LoadResumableData(id string) (map[string]interface{}, error) {

	// use /tmp
	filePath := fmt.Sprint("/tmp/go-gcsproxy-%s.json", id)

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file in LoadResumableData: %v", err)
	}
	defer file.Close() // Ensure the file is closed when the function exits

	// Read the file contents
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading file in LoadResumableData: %v", err)
	}

	// Unmarshal the JSON data
	var dataMap map[string]interface{}
	err = json.Unmarshal(data, &dataMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling ResumableData: %v", err)
	}

	log.Debug(fmt.Sprintf("read ResumableData: %s", data))
	return dataMap, nil
}
