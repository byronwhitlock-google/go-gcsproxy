package main

import (
	"fmt"
	"math/rand"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
)


func generateRandom19DigitNumber() int{

	// Generate the first digit (1-9) to avoid leading zero
	firstDigit := rand.Intn(9) + 1

	// Generate the next 18 digits (0-9)
	var number int64 = int64(firstDigit)
	for i := 0; i < 18; i++ {
		number = number*10 + int64(rand.Intn(10))
	}

	fmt.Printf("Random 19-digit number: %019d\n", number)
	return int(number)
}

func generateHeadersList(f *proxy.Flow) (map[string]string,string){
	defaultMap := map[string]string{
		"Accept-Encoding":  "gzip, deflate",
		"Accept": "application/json",
		"Connection": "keep-alive",
		"Content-Length":"0",
		"Content-Type":"",
		"X-Goog-Api-Client":"cred-type/u",
	}
	boundary_value:=generateRandom19DigitNumber()
	defaultMap["Content-Length"]=strconv.Itoa(len(f.Request.Body))
	defaultMap["Content-Type"]="multipart/related; boundary='==============="+strconv.Itoa(boundary_value)+"=='"
	boundary:="==============="+strconv.Itoa(boundary_value)+"=="
	return defaultMap,boundary
}

//f.Request.URL.Path
//"/download/storage/v1/b/ehorning-axlearn/o/README.md"
func getBucketNameFromRequestUri(urlPath string)string{

	//Splits for the filepath with b in between
	arr :=strings.Split(urlPath, "/b/") //["/download/storage/v1/","ehorning-axlearn/o/README.md"]
	
	//Splits for the filepath with o in between to get exact path
	res :=strings.Split(arr[1], "/o") // ["ehorning-axlearn/","/README.md"]
	
	// Adding this because there might be a path for bucket, so grabbing only bucket name
	bucketName := strings.Split(res[0],"/")[0] 
	fmt.Println("In BucketName Simple Download")
	fmt.Println(bucketName)
	return bucketName
}
//f.Request.URL.Path
func generateMetadata(f *proxy.Flow, contentType string, objectName string)(map[string]interface{}){

	// var gcsMetadata interface{}
	// gcsMetadataMap, ok := gcsMetadata.(map[string]interface{})
	// if !ok {
	// 	return fmt.Errorf("error: JSON data is not a map")
	// }
	// if gcsMetadataMap["metadata"] == nil {
	// 	gcsMetadataMap["metadata"] = make(map[string]interface{})
	// }


	defaultMap := map[string]interface{}{
		"bucket":getBucketNameFromRequestUri(f.Request.URL.Path),
		"contentType":contentType,
		"name":objectName,
		"metadata":map[string]interface{}{
			"x-unencrypted-content-length":len(f.Request.Body),
			"x-md5Hash":base64_md5hash(f.Request.Body),
		},
			
	}
	return defaultMap
}

func CreateFirstMultipartMimeHeader() textproto.MIMEHeader {
	// Process the part, get header , part value
	mimeHeader := textproto.MIMEHeader{}
	//Content-Type: application/json\nMIME-Version: 1.0
	defaultMap:=map[string]string{
		"Content-Type":"application/json",
		"MIME-Version":"1.0",
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
	defaultMap:=map[string]string{
		"Content-Type":contentType,
		"MIME-Version":"1.0",
		"Content-Transfer-Encoding":"binary",
	}
	//Loop through Map
	for k, v := range defaultMap {
		mimeHeader.Set(k, v)
	}
	return mimeHeader
}

