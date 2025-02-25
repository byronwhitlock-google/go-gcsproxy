/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

func HandleMetadataRequest(f *proxy.Flow) error {

	log.Debug(fmt.Sprintf("HandleMetadataRequest got query string  %s", f.Request.URL.RawQuery))

	queryString := f.Request.URL.Query()

	// we delete all fields because there is no way to reliable filter without getting  on new  on new objects
	queryString.Del("fields")
	f.Request.URL.RawQuery = queryString.Encode()

	log.Debug(fmt.Sprintf("formatted query string to %s", f.Request.URL.RawQuery))
	return nil
}

func HandleMetadataResponse(f *proxy.Flow) error {

	log.Debug(fmt.Sprintf("got metadata response: %s", f.Response.Body))

	// Unmarshal the json contents of the first part.
	var gcsMetadataMap map[string]interface{}
	err := json.Unmarshal(f.Response.Body, &gcsMetadataMap)
	if err != nil {
		return fmt.Errorf("error unmarshalling gcsObjectMetadata: %v", err)
	}

	customMetadata, ok := gcsMetadataMap["metadata"].(map[string]interface{})
	if ok {
		// overwrite the size & hash parameter with the unencrypted size & hash
		gcsMetadataMap["size"] = customMetadata["x-unencrypted-content-length"]
		gcsMetadataMap["md5Hash"] = customMetadata["x-md5Hash"]

		// Now write the gcs object metadata back to the multipart writer
		jsonData, err := json.MarshalIndent(gcsMetadataMap, "", "\t")
		if err != nil {
			return fmt.Errorf("error marshalling gcsObjectMetadata: %v", err)
		}
		f.Response.Body = jsonData
		log.Debug(fmt.Sprintf("rewrote metadata response: %s", f.Response.Body))
	}

	return nil
}
