/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package util

/*
	This file provides functionality for resumable upload with subsequent PUT requests(a.k.a chunked upload)
	It's currenly not used
*/
import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/storage"
	cfg "github.com/byronwhitlock-google/go-gcsproxy/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

func parseBearerToken(authHeader string) (string, error) {

	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	return parts[1], nil
}

func updateGcsMetadata(ctx context.Context, authHeader string, bucketName string, objectName string, unencryptedContentLength string, md5Hash string) error {

	bearerToken, err := parseBearerToken(authHeader)
	if err != nil {
		return fmt.Errorf("error parsing bearer token:%v", err)
	}

	// lets use the google SDK so we get some error handling and such.
	// Create a new storage client with the bearer token
	log.Debugf("updating  gs://%v/%v metadata.", bucketName, objectName)

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: bearerToken})
	client, err := storage.NewClient(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	// Get a handle to the object
	obj := client.Bucket(bucketName).Object(objectName)

	// Update the object's metadata
	objectAttrsToUpdate := storage.ObjectAttrsToUpdate{
		Metadata: map[string]string{
			"x-unencrypted-content-length": unencryptedContentLength,
			"x-md5Hash":                    md5Hash,
			"x-encryption-key":             GetKMSKeyName(bucketName),
			"x-proxy-version":              cfg.GlobalConfig.GCSProxyVersion, // TODO: Change this to the global Version in the main package ASAP
		},
	}
	if _, err := obj.Update(ctx, objectAttrsToUpdate); err != nil {
		return fmt.Errorf("failed to update object metadata: %v", err)
	}
	log.Debug("Object metadata updated successfully.")
	return nil
}


func ReadGcsMetadata(ctx context.Context, bucketName string, objectName string) (string,error) {

	// lets use the google SDK so we get some error handling and such.
	// Create a new storage client with the bearer token
	log.Debugf("updating  gs://%v/%v metadata.", bucketName, objectName)
	
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "",fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	// Get a handle to the object
	obj := client.Bucket(bucketName).Object(objectName)

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get object attributes: %v", err)
	}
	log.Debug("Object metadata updated successfully.")
	return attrs.Metadata["x-encryption-key"], nil
}
