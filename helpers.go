package main

import (
	"fmt"
	"strings"
)


func getKMSKeyName(bucketName string) string{

	bucketMap := bucketKeyMappings(config.KmsBucketKeyMapping)

	if bucketMap==nil{
		fmt.Println("In Nil value")
		return ""
	}
	fmt.Println("bucketMap")
	fmt.Println(bucketMap)
	if value, exists := bucketMap[bucketName]; exists {
		fmt.Println(" KMS Key entry exists with value:", value)
		return value
	} else {
		fmt.Println("KMS key entry does not exist")
		return ""
	}
	
}

func getBucketNameMultipartUpload(bucketNameMap map[string]interface{}) string{
	var bucketNamePath string

	for key, value := range bucketNameMap {
			
			if key=="bucket"{
				bucketNamePath=fmt.Sprintf("%s",value)
			}

		}
	bucketName:= strings.Split(bucketNamePath,"/")[0] // may be add gs:// depending on the map object

	fmt.Println("In BucketName Multipart Upload")
	fmt.Println(bucketName)
	return bucketName
}

//f.Request.URL.Path
//"/download/storage/v1/b/ehorning-axlearn/o/README.md"
func getBucketNameSimpleDownload(urlPath string)string{

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