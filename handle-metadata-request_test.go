package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/url"
// 	"reflect"
// 	"testing"

// 	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
// )

// func TestHandleMetadataRequest(t *testing.T) {
// 	f := &proxy.Flow{
// 		Request: &proxy.Request{
// 			URL: &url.URL{
// 				RawQuery: "fields=kind,items(name,size),nextPageToken&alt=json",
// 			},
// 		},
// 	}

// 	err := HandleMetadataRequest(f)
// 	if err != nil {
// 		t.Errorf("HandleMetadataRequest returned an error: %v", err)
// 	}

// 	expectedQuery := "alt=json" // Expecting only alt=json, fields is removed.

// 	if f.Request.URL.RawQuery != expectedQuery {
// 		t.Errorf("RawQuery mismatch.\nExpected: %s\nGot: %s", expectedQuery, f.Request.URL.RawQuery)
// 	}
// }

// func TestHandleMetadataResponse(t *testing.T) {
// 	// Example of a gcsMetadataMap
// 	gcsMetadataMap := map[string]interface{}{
// 		"kind": "storage#objects",
// 		"metadata": map[string]interface{}{
// 			"x-unencrypted-content-length": "12345",
// 			"x-md5Hash":                    "testmd5",
// 		},

// 	}

// 	jsonData, err := json.Marshal(gcsMetadataMap)

// 	if err != nil {
// 		fmt.Errorf("error %v", err)
// 	}

// 	f := &proxy.Flow{
// 		Response:&proxy.Response{
// 			Body:jsonData,
// 		},

// 	}

// 	err = HandleMetadataResponse(f)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	var result map[string]interface{}
// 	err = json.Unmarshal(f.Response.Body, &result)

// 	if err!=nil{
// 		t.Fatal(err)
// 	}

// 	expectedSize :=  gcsMetadataMap["metadata"].(map[string]interface{})["x-unencrypted-content-length"]

// 	expectedMd5Hash := gcsMetadataMap["metadata"].(map[string]interface{})["x-md5Hash"]

// 	if !reflect.DeepEqual(result["size"], expectedSize) {
// 		t.Errorf("Size mismatch.\nExpected: %v\nGot: %v", expectedSize, result["size"])
// 	}
// 	if !reflect.DeepEqual(result["md5Hash"], expectedMd5Hash) {
// 		t.Errorf("Size mismatch.\nExpected: %v\nGot: %v", expectedMd5Hash, result["md5Hash"])
// 	}
// }

// //Missing metadata, negative test
// func TestHandleMetadataResponseMissingMetadata(t *testing.T) {
// 	// Example of a gcsMetadataMap without "metadata"
// 	gcsMetadataMap := map[string]interface{}{
// 		"kind": "storage#objects",
// 		"size":"123",
// 	}

// 	jsonData, err := json.Marshal(gcsMetadataMap)
// 	if err != nil {
// 		fmt.Errorf("error %v", err)
// 	}

// 	f := &proxy.Flow{
// 		Response: &proxy.Response{
// 			Body: jsonData,
// 		},
// 	}

// 	err = HandleMetadataResponse(f)
// 	if err == nil {
// 		t.Error("Expected an error but got nil")
// 	}
// }
