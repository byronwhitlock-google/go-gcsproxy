/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package handlers

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"cloud.google.com/go/storage"
)

// streamFileUpload uploads an object via a stream.
func StreamFileUpload(file multipart.File, bucket, object string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	wc.ChunkSize = 0 // note retries are not supported for chunk size 0.

	if _, err = io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	// TODO: Fix code to do error handling and allow closing the writer.
	// error while closing the writer - Writer.Close: json: invalid use of ,string struct tag, trying to unmarshal unquoted value into uint64

	// Data can continue to be added to the file until the writer is closed.
	// if err := wc.Close(); err != nil {
	// 	return fmt.Errorf("Writer.Close: %w", err)
	// }
	wc.Close()

	return err
}
