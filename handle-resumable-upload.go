package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
)

// this is the raw data to be encoded.
func HandleResumablePutRequest(f *proxy.Flow) error {
	/*
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
	*/
	// update content range
	// content-range: bytes 0-72355493/72355494

	unencryptedFileContent := bytes.NewBuffer(f.Request.Body)

	// Encrypt the intercepted file
	encryptedData, err := encryptBytes(f.Request.Raw().Context(),
		config.KmsResourceName,
		unencryptedFileContent.Bytes())

	if err != nil {
		return fmt.Errorf("error encrypting  request: %v", err)
	}

	byteRangeHeader := f.Request.Header.Get("content-range")
	start, end, size := parseByteRangeHeader(byteRangeHeader)

	if !(start == 0 && end == size) {
		return fmt.Errorf("unsupported Byte range detected '%v'", byteRangeHeader)
	}

	// rewrite the byte range header to what we have already enxcrypted...
	size = bytes.Count(encryptedData, []byte{}) - 1
	end = size
	start = 0

	byteRangeHeader = fmt.Sprintf("bytes %v-%v/%v", start, end, size)
	f.Request.Header.Set("content-range", byteRangeHeader)
	return nil
}

// rangeString = "bytes 0-72355493/72355494"
func parseByteRangeHeader(rangeString string) (start int, end int, size int) {
	parts := strings.Split(rangeString, "/")
	if len(parts) != 2 {
		return 0, 0, 0 // Invalid range format
	}

	size, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, 0 // Error parsing size
	}

	rangeParts := strings.Split(strings.TrimSpace(parts[0]), "-")
	if len(rangeParts) != 2 {
		return 0, 0, 0 // Invalid range format
	}

	start, err = strconv.Atoi(strings.TrimSpace(rangeParts[0]))
	if err != nil {
		return 0, 0, 0 // Error parsing start
	}

	end, err = strconv.Atoi(strings.TrimSpace(rangeParts[1]))
	if err != nil {
		return 0, 0, 0 // Error parsing end
	}

	return start, end, size
}

/*
func HandleResumablePostRequest(f *proxy.Flow) error {
	return nil //do nothing
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
*/
