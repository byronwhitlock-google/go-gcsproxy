package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"

	log "github.com/sirupsen/logrus"
)

// ParseMultipartRequest parses a multipart request like the one provided.
func ParseMultipartRequest(reader io.Reader, boundary string) (string, string, error) {
	// Create a multipart reader with the given boundary.
	mr := multipart.NewReader(reader, boundary)

	// Parse the first part (JSON metadata).
	part, err := mr.NextPart()
	if err != nil {
		return "", "", fmt.Errorf("error reading first part: %w", err)
	}

	// Decode the JSON metadata.
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, part)
	if err != nil {
		return "", "", fmt.Errorf("error reading file content: %w", err)
	}
	var gcsMetadata string
	gcsMetadata = string(buf.Bytes())

	//var gcsMetadata map[string]interface{}
	//json.Unmarshal(jsonBuf.Bytes(), &gcsMetadata)

	// Parse the second part (file content).
	part, err = mr.NextPart()
	if err != nil {
		return "", "", fmt.Errorf("error reading second part: %w", err)
	}
	// Decode the JSON metadata.
	buf = new(bytes.Buffer)
	_, err = io.Copy(buf, part)
	if err != nil {
		return "", "", fmt.Errorf("error reading file content: %w", err)
	}
	var fileContent string
	fileContent = string(buf.Bytes())

	return gcsMetadata, fileContent, nil
}


func ParseMultipartRequest1(reader io.Reader, boundary string) (string, string, error) {
	mr := multipart.NewReader(reader, boundary)

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
		fmt.Println("Processing Header: %s\n", header)
		// Get Value of part 
		if part.FileName() == "" {
			fieldValue, err := io.ReadAll(part)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("File content: %s\n", string(fieldValue))
		}
	}
    return "","",nil

}
