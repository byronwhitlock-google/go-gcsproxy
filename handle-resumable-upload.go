package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strconv"

	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

// this is the raw data to be encoded.
func HandleResumablePutRequest(f *proxy.Flow) error {

	byteRangeHeader := f.Request.Header.Get("Content-Range")
	start, end, size, err := parseByteRangeHeader(byteRangeHeader)
	if err != nil {
		return err
	}

	if !(start == 0 && end+1 == size) {
		return fmt.Errorf("unsupported Byte range detected '%v'", byteRangeHeader)
	}
	// unencryptedFileContent := bytes.NewBuffer(f.Request.Body)

	// // Encrypt the intercepted file
	// encryptedData, err := encryptBytes(f.Request.Raw().Context(),
	// 	config.KmsResourceName,
	// 	unencryptedFileContent.Bytes())

	// if err != nil {
	// 	return fmt.Errorf("error encrypting  request: %v", err)
	// }

	// // rewrite the byte range header to what we have already enxcrypted...
	// size = len(encryptedData) //bytes.Count(encryptedData, []byte{})
	// end = size
	// start = 0

	// newByteRangeHeader := fmt.Sprintf("bytes %v-%v/%v", start, end-1, size)
	// f.Request.Header.Set("Content-Range", newByteRangeHeader)
	// f.Request.Body = encryptedData

	// log.Debug(fmt.Sprint("encrypted PUT request headers: %v", f.Request.Header))
	// //log.Debug(fmt.Sprint("Encrypted PUT request Body: %s", f.Request.Body))

	// // Save the original content length for rewriting when download.
	// f.Request.Header.Set("gcs-proxy-original-content-length",
	// 	f.Request.Header.Get("Content-Length"))

	// f.Request.Header.Set("gcs-proxy-unencrypted-file-size",
	// 	strconv.Itoa(unencryptedFileContent.Len()))

	// // save the original md5 has or gsutil/gcloud will delete after upload if it sees it is different
	// f.Request.Header.Set("gcs-proxy-original-md5-hash",
	// 	base64_md5hash(unencryptedFileContent.Bytes()))

	// first we need the uploader id so we can get the resumable metadata.
	log.Debug(fmt.Sprintf("got query string  %s", f.Request.URL.RawQuery))

	uploadId := f.Request.URL.Query().Get("upload_id")

	if uploadId == "" {
		return fmt.Errorf("missing upload id in query string: %v", f.Request.URL.RawQuery)
	}
	resumeData, err := LoadResumableData(uploadId)
	if err != nil {
		return fmt.Errorf("error Loading Resumable Data: %v", err)
	}

	url, err := url.Parse(fmt.Sprintf("https://storage.googleapis.com/upload/storage/v1/b/%v/o?name=%v", resumeData["bucket"], resumeData["name"]))
	if err != nil {
		panic(err) // Handle the error appropriately in a real application
	}
	f.Request.URL = url

	ConvertSinglePartUploadtoMultiPartUpload(f)

	// don't update metadata until we send the request. otherwise the object will not exist
	// defer updateGcsMetadata(f.Request.Raw().Context(),
	// 	f.Request.Header.Get("Authorization"),
	// 	resumeData["bucket"],
	// 	resumeData["name"],
	// 	strconv.Itoa(unencryptedFileContent.Len()),
	// 	base64_md5hash(unencryptedFileContent.Bytes()))
	return nil
}

func HandleResumablePutResponse(f *proxy.Flow) error {

	// this code should never be executted. after converting a resumable PUT to a POST, it should fail.

	log.Debug("in HandleResumablePutResponse")

	var jsonResponse map[string]interface{}
	// turn the response body into a dynamic json map we can use
	err := json.Unmarshal(f.Response.Body, &jsonResponse)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}
	log.Debug(jsonResponse)

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

func HandleResumablePostRequest(f *proxy.Flow) error {
	// strip X-upload-content-length
	f.Request.Header.Del("x-upload-content-length")
	f.Request.Header.Del("X-Upload-Content-Length")
	return nil //do nothing
}

func HandleResumablePostResponse(f *proxy.Flow) error {

	// the client posts the file name in the request body. store that and other info in our session file.
	log.Debug(fmt.Sprintf("HandleResumablePostResponse Response.Body: %s", f.Request.Body))

	// Unmarshal the json contents of the first part.
	var dataMap map[string]string
	err := json.Unmarshal(f.Request.Body, &dataMap)
	if err != nil {
		return fmt.Errorf("error unmarshalling gcsObjectMetadata in HandleResumablePostResponse: %v", err)
	}

	// uploader id comes from GCS so it is in the Response
	uploaderId := f.Response.Header.Get("X-GUploader-UploadID")
	if uploaderId == "" {
		return fmt.Errorf("missing X-GUploader-UploadID header")
	}

	StoreResumableData(uploaderId, dataMap)

	return nil
}

// writes data to a file by id
func StoreResumableData(id string, dataMap map[string]string) error {

	// use /tmp
	filePath := fmt.Sprintf("/tmp/go-gcsproxy-%s.json", id)

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
func LoadResumableData(id string) (map[string]string, error) {

	// use /tmp
	filePath := fmt.Sprintf("/tmp/go-gcsproxy-%s.json", id)

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

	// now delete the temp file for now.
	// TODO: resumable streams would store partial data here.
	// TODO: implement streaming functions so resumable uploads can cancel with partial data within a request.
	defer os.Remove(filePath)

	// Unmarshal the JSON data
	var dataMap map[string]string
	err = json.Unmarshal(data, &dataMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling ResumableData: %v", err)
	}

	log.Debug(fmt.Sprintf("read ResumableData: %s", data))
	return dataMap, nil
}
