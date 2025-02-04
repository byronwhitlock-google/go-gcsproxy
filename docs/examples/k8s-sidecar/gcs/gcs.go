/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package gcs

import (
	"fmt"
	"go-api/gcs/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func checkError(ctx *gin.Context, err error) bool {
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return true
	}
	return false
}

func DownloadObjects(ctx *gin.Context) {
	gcsBucket := ctx.Param("gcsBucket")
	gcsObject := ctx.Param("gcsObject")
	response := "Endpoint to download " + string(gcsObject) + "object from " + string(gcsBucket) + " GCS bucket."
	fmt.Println(response)

	data, err := handlers.DownloadFileIntoMemory(gcsBucket, gcsObject)
	if checkError(ctx, err) {
		return
	}
	ctx.String(http.StatusOK, string(data))
}

func UploadObjects(ctx *gin.Context) {
	gcsBucket := ctx.Param("gcsBucket")
	gcsObject := ctx.Param("gcsObject")
	response := "Endpoint to upload " + string(gcsObject) + "object to " + string(gcsBucket) + " GCS bucket."
	fmt.Println(response)

	file, err := ctx.FormFile("file")
	if checkError(ctx, err) {
		return
	}

	blobFile, err := file.Open()
	if checkError(ctx, err) {
		return
	}
	err = handlers.StreamFileUpload(blobFile, gcsBucket, gcsObject)

	if checkError(ctx, err) {
		return
	}

	ctx.String(http.StatusOK, response)
}
