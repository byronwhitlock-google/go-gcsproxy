package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"

	log "github.com/sirupsen/logrus"
)

func ParseMultipartRequest(reader io.Reader, boundary string) (*bytes.Buffer, error) {
	mr := multipart.NewReader(reader, boundary)
	// New empty buffer
	body := &bytes.Buffer{}
	num:=0
	// Creates a new multipart Writer with a random boundary, writing to the empty
	// buffer
	writer := multipart.NewWriter(body)

	err2:= writer.SetBoundary(boundary)
	if err2 != nil {
		log.Fatal(err2)
	}
	for {
		// Read the next part
		num+=1
		part, err := mr.NextPart()
		if err == io.EOF {
			// We've reached the end of the multipart data
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		// Process the part, get header , part value
		header := part.Header
		metadataHeader := textproto.MIMEHeader{}
		//Loop through Map 
		for k, v := range header {
			if num == 2 && k =="Content-Type"{
					metadataHeader.Set(k,"application/octet-stream")
			}else {
					metadataHeader.Set(k,v[0])
			}
			
    	}
		
		writer_part, err := writer.CreatePart(metadataHeader)
		if err != nil {
			log.Fatal(err)
		}

		// Get Value of part 
		if part.FileName() == "" {
			fieldValue , err := io.ReadAll(part)
			if err != nil {
				log.Fatal(err)
			}
			// Write/encrypt if needed the part body
			if num == 2{
				encrypted_data ,err := encrypt_tink(fieldValue)
				if err != nil {
					log.Fatal(err)
				}
				writer_part.Write(encrypted_data)
			}else{
				//Change the first part contentType to octet-stream 
				var result map[string]interface{}

					err:=json.Unmarshal(fieldValue, &result)
					if err != nil {
						log.Fatalf("Error unmarshalling JSON: %v", err)
					}
					fmt.Println(result)

					result["contentType"]="application/octet-stream"

					jsonData, err := json.Marshal(result)
					if err != nil {
						fmt.Println("Error marshaling to JSON:", err)
					}

					writer_part.Write([]byte(string(jsonData)))

			}
			
		}
		
	}
	writer.Close()
	fmt.Println(body)
    return body,nil

}
