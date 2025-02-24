/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	rawLog "log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	cfg "github.com/byronwhitlock-google/go-gcsproxy/config"
	"github.com/byronwhitlock-google/go-gcsproxy/crypto"
	gcsproxy "github.com/byronwhitlock-google/go-gcsproxy/proxy"
	"go.opentelemetry.io/otel/metric"

	log "github.com/sirupsen/logrus"
)

// makefile will turn this into a version
var Version = ".3"

func main() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	go func() {
		s := <-sigc
		log.Info("Signal Caught: ", s)
		os.Exit(0)
	}()

	otelEnabled := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	// If OTEL is configured. Setup the custom metrics to capture encrypt/decrypt time.
	if otelEnabled != "" {
		initMetrics()
		initConfig()
		runner := gcsproxy.NewProxyRunner(cfg.GlobalConfig)

		// Setup metrics, tracing, and context propagation
		ctx := context.Background()
		shutdown, err := setupOpenTelemetry(ctx)
		if err != nil {
			log.Fatalf("Error setting up OpenTelemetry. Error:", err)
		}

		// Start the GCS proxy server, and shutdown and flush telemetry after it exits.
		slog.InfoContext(ctx, "server starting...")
		if err = errors.Join(runner.Start(), shutdown(ctx)); err != nil {
			log.Fatalf("Server exited with error. Error:", err)
		}
	} else {
		initConfig()
		runner := gcsproxy.NewProxyRunner(cfg.GlobalConfig)
		err := runner.Start()
		if err != nil {
			log.Fatalf("Fatal error to start the GCS proxy. Error:", err)
		} else {
			log.Info("GCS proxy started successfully")
		}
	}
}

func initMetrics() {
	var err error
	crypto.EncryptTime, err = crypto.Meter.Float64Gauge(
		"proxy.encryptTime",
		metric.WithDescription("GCS Proxy Encryption time"),
		metric.WithUnit("seconds"),
	)
	if err != nil {
		panic(err)
	}

	crypto.DecryptTime, err = crypto.Meter.Float64Gauge(
		"proxy.decryptTime",
		metric.WithDescription("GCS Proxy Decryption time"),
		metric.WithUnit("seconds"),
	)
	if err != nil {
		panic(err)
	}
}

func initConfig() {
	config := cfg.LoadConfig()

	if config.Version {
		log.Infof("go-gcsproxy: %v", Version)
		usage()
		os.Exit(0)
	}

	if config.Debug > 0 {
		rawLog.SetFlags(rawLog.LstdFlags | rawLog.Lshortfile)
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	if config.Debug == 2 {
		log.SetReportCaller(true)
	}
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	err := checkKmsBucketKeyMapping()
	if err != nil {
		log.Fatalf("\n>>> unable to initialize KmsBucketKeyMapping. %v", err)
	}

	configJson, _ := json.MarshalIndent(config, "", "\t")
	log.Infof("go-gcsproxy version '%v' Startting... %v", config.Version, string(configJson))
}

func usage() {
	flag.Usage()
	fmt.Println("\nEnvironment variables supported:")
	fmt.Println("  PROXY_CERT_PATH")
	fmt.Println("  SSL_INSECURE")
	fmt.Println("  DEBUG_LEVEL")
	fmt.Println("  GCP_KMS_BUCKET_KEY_MAPPING")
}

func checkKmsBucketKeyMapping() error {
	var ctx = context.TODO()
	bucketKeyMap := cfg.GlobalConfig.KmsBucketKeyMapping
	if bucketKeyMap == nil {
		return fmt.Errorf("No KmsBucketKeyMapping found")
	}
	for _, value := range bucketKeyMap {
		_, err := crypto.EncryptBytes(ctx, value, []byte("Hello, World!"))
		if err != nil {
			return err
		}
	}
	return nil
}
