package main

import (
	"fmt"
	"strings"
)


func getKMSKeyName(bucketName string) string{

	bucketMap := bucketKeyMappings(config.KmsBucketKeyMapping)

	if bucketMap==nil{
		fmt.Println("In Nil value")
		return config.KmsResourceName
	}
	fmt.Println("bucketMap")
	fmt.Println(bucketMap)
	if value, exists := bucketMap[bucketName]; exists {
		fmt.Println("Key exists with value:", value)
		return value
	} else {
		fmt.Println("Key 'city' does not exist")
		return config.KmsResourceName
	}
	
}

func getBucketName(bucketNameMap map[string]interface{}) string{
	var bucketName string
	var fileName string

	for key, value := range bucketNameMap {
			
			if key=="bucket"{
				bucketName=fmt.Sprintf("%s",value)
			}
			if key=="name"{
				fileName=fmt.Sprintf("%s",value)
			}

		}
	fullPath:= bucketName + "/"+fileName // may be add gs:// depending on the map object

	fmt.Println("In full path")
	fmt.Println(fullPath)
	return fullPath
}

//f.Request.URL.Path
//"/download/storage/v1/b/ehorning-axlearn/o/README.md"
func getBucketNameSimpleDownload(urlPath string)string{

	//Splits for the filepath with b in between
	arr :=strings.Split(urlPath, "b") //["/download/storage/v1/","ehorning-axlearn/o/README.md"]
	
	//Splits for the filepath with o in between to get exact path
	res :=strings.Split(arr[1], "/o/") // ["ehorning-axlearn/","/README.md"]

	fullPath := res[0][1:]+"/"+res[1]
	fmt.Println(fullPath)
	return fullPath
}