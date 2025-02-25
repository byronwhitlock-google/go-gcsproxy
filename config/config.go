/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package cfg

import (
	"flag"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Config struct {
	Version bool // show version

	Addr        string // proxy listen addr
	WebAddr     string // web interface listen addr
	SslInsecure bool   // not verify upstream server SSL/TLS certificates.

	CertPath string // path of generate cert files
	Debug    int    // debug mode: 1 - print debug log, 2 - show debug from

	Dump      string // dump filename
	DumpLevel int    // dump level: 0 - header, 1 - header + body

	// kms options
	kmsBucketKeyMappingString string
	KmsBucketKeyMapping       map[string]string

	Upstream        string // upstream proxy
	UpstreamCert    bool   // Connect to upstream server to look up certificate details. Default: True
	EncryptDisabled bool
	GCSProxyVersion string
}

var GlobalConfig *Config // Global variable

func LoadConfig() *Config {
	config := new(Config)
	config.EncryptDisabled = isEncryptDisabled()

	defaultSslInsecure := envConfigBoolWithDefault("SSL_INSECURE", true)
	defaultCertPath := envConfigStringWithDefault("PROXY_CERT_PATH", "/Users/lkolluru/working-dir/apple/go-gcsproxy/test")
	defaultDebug := envConfigIntWithDefault("DEBUG_LEVEL", 0)
	defaultKmsBucketKeyMappingString := envConfigStringWithDefault("GCP_KMS_BUCKET_KEY_MAPPING", "*:projects/cmetestproj/locations/global/keyRings/gcsproxytest/cryptoKeys/gcsproxy/cryptoKeyVersions/2")

	flag.BoolVar(&config.Version, "version", false, "show go-gcsproxy version")
	flag.StringVar(&config.Addr, "port", ":9080", "proxy listen addr")
	flag.StringVar(&config.WebAddr, "web_port", ":9081", "web interface listen addr")
	flag.BoolVar(&config.SslInsecure, "ssl_insecure", defaultSslInsecure, "don't verify upstream server SSL/TLS certificates.")

	flag.StringVar(&config.CertPath, "cert_path", defaultCertPath, "path to cert. if 'mitmproxy-ca.pem' is not present here, it will be generated.")
	flag.IntVar(&config.Debug, "debug", defaultDebug, "debug level: 0 - ERROR, 1 - DEBUG, 2 - TRACE")
	flag.StringVar(&config.Dump, "dump", "", "filename to dump req/responses for debugging")
	flag.IntVar(&config.DumpLevel, "dump_level", 0, "dump level: 0 - header, 1 - header + body")
	flag.StringVar(&config.Upstream, "upstream", "", "upstream proxy")
	// "*:global-key" or "bucket/path:project/key,bucket2:key2" but the global key overrides all the other keys
	flag.StringVar(&config.kmsBucketKeyMappingString, "kms_bucket_key_mappings", defaultKmsBucketKeyMappingString, "Maps Bucket name to KMS keys. Proxy encrypts object uploaded to BUCKET with KEY stored in KMS. Setting BUCKET to * will encrypt/decrypt all GCS calls. Format is `BUCKET:KEY1,BUCKET2:KEY2` for example: `mygcsbucket:projects/<project_id>/locations/<global|region>/keyRings/<key_ring>/cryptoKeys/<key>`")

	flag.BoolVar(&config.UpstreamCert, "upstream_cert", false, "connect to upstream server to look up certificate details")
	flag.Parse()
	config.KmsBucketKeyMapping = getBucketKeyMappings(config.kmsBucketKeyMappingString)
	config.GCSProxyVersion = "0.2"
	GlobalConfig = config
	return config
}

// Parsing the "*:global-key" or "bucket/path:project/key,bucket2:key2" but the global key overrides all the other keys
func getBucketKeyMappings(bucketKeyMapString string) map[string]string {

	if bucketKeyMapString == "" {
		log.Debug("No Bucket Key Mapping given")
		return nil
	}

	bucketKeyMap := make(map[string]string)
	bucketKeys := strings.Split(bucketKeyMapString, ",")
	for i := 0; i < len(bucketKeys); i++ {

		bucketKeyArray := strings.Split(bucketKeys[i], ":")
		bucketKeyMap[bucketKeyArray[0]] = bucketKeyArray[1]
	}

	log.Debugf("BucketkeyMapping: %v", bucketKeyMap)
	return bucketKeyMap

}

func isEncryptDisabled() bool {
	if os.Getenv("GCS_PROXY_DISABLE_ENCRYPTION") == "" {
		return false
	}
	return true
}

func envConfigStringWithDefault(key string, defValue string) string {
	envVar := os.Getenv(key)
	if len(envVar) == 0 {
		return defValue
	}
	return envVar
}

func envConfigBoolWithDefault(key string, defValue bool) bool {
	envVar, boolError := strconv.ParseBool(os.Getenv(key))
	if boolError == nil {
		return envVar
	}
	return defValue
}

func envConfigIntWithDefault(key string, defValue int) int {
	envVar, intError := strconv.Atoi(os.Getenv(key))
	if intError == nil {
		return envVar
	}
	return defValue
}
