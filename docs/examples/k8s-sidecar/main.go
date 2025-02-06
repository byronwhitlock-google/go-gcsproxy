/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package main

import (
	"context"
	"errors"
	"fmt"
	"go-api/gcs"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const (
	// DebugMode indicates gin mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates gin mode is release.
	ReleaseMode = "release"
	// TestMode indicates gin mode is test.
	TestMode = "test"
)

func runServer() error {
	// Set GIN Mode
	gin.SetMode(DebugMode)

	// Initialize Gin router and set router configs
	router := gin.Default()
	router.Use(otelgin.Middleware("go-api-gcsapi"))
	router.ContextWithFallback = true
	router.SetTrustedProxies(nil)

	// Define GET method to upload GCS object
	router.GET("/download/:gcsBucket/:gcsObject", gcs.DownloadObjects)

	// POST method to download GCS object
	router.POST("/upload/:gcsBucket/:gcsObject", gcs.UploadObjects)

	// Run Gin router
	return router.Run(":8080")
}

func main() {

	// uncomment for testing locally
	// if err := os.Setenv("https_proxy", "http://127.0.0.1:9080"); err != nil {
	// 	fmt.Println("Error setting environment variable:", err)
	// 	return
	// }
	// if err := os.Setenv("REQUESTS_CA_BUNDLE", "/proxy/certs/mitmproxy-ca-cert.pem"); err != nil {
	// 	fmt.Println("Error setting environment variable:", err)
	// 	return
	// }
	// if err := os.Setenv("SSL_CERT_FILE", "/proxy/certs/mitmproxy-ca-cert.pem"); err != nil {
	// 	fmt.Println("Error setting environment variable:", err)
	// 	return
	// }

	ctx := context.Background()

	// Setup logging
	setupLogging()

	// Setup metrics, tracing, and context propagation
	shutdown, err := setupOpenTelemetry(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "error setting up OpenTelemetry", slog.Any("error", err))
		os.Exit(1)
	}

	// Run the http server, and shutdown and flush telemetry after it exits.
	slog.InfoContext(ctx, "server starting...")
	if err = errors.Join(runServer(), shutdown(ctx)); err != nil {
		slog.ErrorContext(ctx, "server exited with error", slog.Any("error", err))
		os.Exit(1)
	}

}
