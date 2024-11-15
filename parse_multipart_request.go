package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
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
