/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package util

import (
	"fmt"
	"math/rand"
	"net/textproto"
	"strconv"
	"strings"

	cfg "github.com/byronwhitlock-google/go-gcsproxy/config"
	"github.com/byronwhitlock-google/go-gcsproxy/crypto"
	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

func GetKMSKeyName(bucketName string) string {

	bucketMap := cfg.GlobalConfig.KmsBucketKeyMapping

	if bucketMap == nil {
		log.Debug("No bucket mapping found")
		return ""
	}

	// Global key is highest priority
	if value, exists := bucketMap["*"]; exists {
		log.Debugf("Global KMS Key entry exists with value: %v", value)
		return value
	}
	// If Global key , then check other bucket to KMS key mapping
	if value, exists := bucketMap[bucketName]; exists {
		log.Debugf(" KMS Key entry exists with value: %v", value)
		return value
	} else {
		log.Debug("KMS key entry does not exist")
		return ""
	}

}

func GetBucketNameFromGcsMetadata(bucketNameMap map[string]interface{}) string {
	var bucketNamePath string

	for key, value := range bucketNameMap {

		if key == "bucket" {
			bucketNamePath = fmt.Sprintf("%s", value)
		}

	}
	bucketName := strings.Split(bucketNamePath, "/")[0]

	log.Debugf("In Multipart Upload for bucket name: %v", bucketName)
	return bucketName
}

func GenerateHeadersList(f *proxy.Flow) (map[string]string, string) {
	defaultMap := map[string]string{
		"Accept-Encoding":   "gzip, deflate",
		"Accept":            "application/json",
		"Connection":        "keep-alive",
		"Content-Length":    "0",
		"Content-Type":      "",
		"X-Goog-Api-Client": "cred-type/u",
	}
	boundary_value := generateRandom19DigitNumber()
	defaultMap["Content-Length"] = strconv.Itoa(len(f.Request.Body))
	defaultMap["Content-Type"] = "multipart/related; boundary='===============" + strconv.Itoa(boundary_value) + "=='"
	boundary := "===============" + strconv.Itoa(boundary_value) + "=="
	return defaultMap, boundary
}

// f.Request.URL.Path
// "/download/storage/v1/b/ehorning-axlearn/o/README.md"
// "/bucket-name/object-path"
func GetBucketNameFromRequestUri(urlPath string) string {
	var bucketName string
	if strings.Contains(urlPath, "/b/") {
		//Splits for the filepath with b in between
		arr := strings.Split(urlPath, "/b/") //["/download/storage/v1/","ehorning-axlearn/o/README.md"]

		//Splits for the filepath with o in between to get exact path
		res := strings.Split(arr[1], "/o") // ["ehorning-axlearn/","/README.md"]

		// Adding this because there might be a path for bucket, so grabbing only bucket name
		bucketName = strings.Split(res[0], "/")[0]
	} else {
		// handle path=/bucket-name/object-path
		bucketName = strings.Split(urlPath, "/")[1]
	}
	log.Debugf("getBucketNameFromRequestUri bucketName: %v", bucketName)
	return bucketName
}

func GetObjectNameFromRequestUri(urlPath string) string {
	var objectName string
	if strings.Contains(urlPath, "/o/") {
		//Splits for the filepath with b in between
		arr := strings.Split(urlPath, "/o/") //["/download/storage/v1/ehorning-axlearn","README.md"]
		// Adding this because there might be a path for bucket, so grabbing only bucket name
		objectName = arr[1]
	} else {
		// handle path=/bucket-name/object-path
		objectName = ""
	}
	log.Debugf("GetObjectNameFromRequestUri objectName: %v", objectName)
	return objectName
}

// TODO: move this back to handle-singlepart-upload for clarity
func GenerateMetadata(f *proxy.Flow, contentType string, objectName string) map[string]interface{} {
	bucketName := GetBucketNameFromRequestUri(f.Request.URL.Path)
	defaultMap := map[string]interface{}{
		"bucket":      bucketName,
		"contentType": contentType,
		"name":        objectName,
		"metadata": map[string]interface{}{
			"x-unencrypted-content-length": len(f.Request.Body),
			"x-md5Hash":                    crypto.Base64MD5Hash(f.Request.Body),
			"x-encryption-key":             GetKMSKeyName(bucketName),
			"x-proxy-version":              cfg.GlobalConfig.GCSProxyVersion, // TODO: Change this to the global Version in the main package ASAP
		},
	}
	return defaultMap
}

func CreateFirstMultipartMimeHeader() textproto.MIMEHeader {
	// Process the part, get header , part value
	mimeHeader := textproto.MIMEHeader{}
	//Content-Type: application/json\nMIME-Version: 1.0
	defaultMap := map[string]string{
		"Content-Type": "application/json",
		"MIME-Version": "1.0",
	}
	//Loop through Map
	for k, v := range defaultMap {
		mimeHeader.Set(k, v)
	}
	return mimeHeader
}

func CreateSecondMultipartMimeHeader(contentType string) textproto.MIMEHeader {
	// Process the part, get header , part value
	mimeHeader := textproto.MIMEHeader{}
	//Content-Type: text/markdown\nMIME-Version: 1.0\nContent-Transfer-Encoding: binary
	defaultMap := map[string]string{
		"Content-Type":              contentType,
		"MIME-Version":              "1.0",
		"Content-Transfer-Encoding": "binary",
	}
	//Loop through Map
	for k, v := range defaultMap {
		mimeHeader.Set(k, v)
	}
	return mimeHeader
}

func generateRandom19DigitNumber() int {

	// Generate the first digit (1-9) to avoid leading zero
	firstDigit := rand.Intn(9) + 1

	// Generate the next 18 digits (0-9)
	var number int64 = int64(firstDigit)
	for i := 0; i < 18; i++ {
		number = number*10 + int64(rand.Intn(10))
	}

	return int(number)
}
