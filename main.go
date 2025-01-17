/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	rawLog "log"
	"os"
	"os/signal"
	"syscall"

	cfg "github.com/byronwhitlock-google/go-gcsproxy/config"
	"github.com/byronwhitlock-google/go-gcsproxy/crypto"
	gcsproxy "github.com/byronwhitlock-google/go-gcsproxy/proxy"

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

	initConfig()
	runner := gcsproxy.NewProxyRunner(cfg.GlobalConfig)
	err := runner.Start()
	if err != nil {
		log.Fatalf("Fatal error to start the GCS proxy. Error:", err)
	} else {
		log.Info("GCS proxy started successfully")
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
