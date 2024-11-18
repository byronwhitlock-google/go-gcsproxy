package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/textproto"

	log "github.com/sirupsen/logrus"
)

func ParseMultipartRequest(reader io.Reader, boundary string) (*bytes.Buffer, error) {
	mr := multipart.NewReader(reader, boundary)
	// New empty buffer
	body := &bytes.Buffer{}
	// Creates a new multipart Writer with a random boundary, writing to the empty
	// buffer
	writer := multipart.NewWriter(body)

	err2:= writer.SetBoundary(boundary)
	if err2 != nil {
		log.Fatal(err2)
	}
	for {
		// Read the next part
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
			metadataHeader.Set(k,v[0])
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
			// Write the part body
			writer_part.Write([]byte(fieldValue))
			
		}
		
	}
	writer.Close()
    return body,nil

}
