package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	rawLog "log"
	"os"

	"github.com/byronwhitlock-google/go-mitmproxy/addon"
	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	"github.com/byronwhitlock-google/go-mitmproxy/web"
	log "github.com/sirupsen/logrus"
)

// makefile will turn this into a version
var Version = ".3"

type Config struct {
	version bool // show version

	Addr        string // proxy listen addr
	WebAddr     string // web interface listen addr
	SslInsecure bool   // not verify upstream server SSL/TLS certificates.

	CertPath string // path of generate cert files
	Debug    int    // debug mode: 1 - print debug log, 2 - show debug from

	Dump      string // dump filename
	DumpLevel int    // dump level: 0 - header, 1 - header + body

	// kms options
	KmsResourceName string

	Upstream     string // upstream proxy
	UpstreamCert bool   // Connect to upstream server to look up certificate details. Default: True
}

// global config variable
var config *Config

func main() {
	config = loadConfig()
	if config.version {
		fmt.Println("go-gcsproxy: " + Version)
		Usage()
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

	if config.KmsResourceName == "" {
		fmt.Printf("\n>>> kms_resource_name empty.\n")
		Usage()
		os.Exit(0)
	}

	err := CheckKMS()
	if err != nil {
		fmt.Printf("\n>>> unable to initialize Google KMS. %v", err)
		os.Exit(0)
	}

	opts := &proxy.Options{
		Debug:             config.Debug,
		Addr:              config.Addr,
		StreamLargeBodies: 1024 * 1024 * 1024 * 64, // TODO: we need to implement streaming intercept functions set to 64GB for now!
		SslInsecure:       config.SslInsecure,
		CaRootPath:        config.CertPath,
		Upstream:          config.Upstream,
	}

	p, err := proxy.NewProxy(opts)
	if err != nil {
		log.Fatal(err)
	}

	if !config.UpstreamCert {
		p.AddAddon(proxy.NewUpstreamCertAddon(false))
		log.Infoln("UpstreamCert config false")
	}

	p.AddAddon(&proxy.LogAddon{})
	p.AddAddon(web.NewWebAddon(config.WebAddr))

	p.AddAddon(&EncryptGcsPayload{})
	p.AddAddon(&DecryptGcsPayload{})
	p.AddAddon(&GetReqHeader{})

	if config.Dump != "" {
		dumper := addon.NewDumperWithFilename(config.Dump, config.DumpLevel)
		p.AddAddon(dumper)
	}

	configJson, _ := json.MarshalIndent(config, "", "\t")
	msg := fmt.Sprintf("go-gcsproxy version '%v' Started. %v", config.version, string(configJson))
	log.Info(msg)
	log.Info(fmt.Sprintf("Encryption enabled: %t", !IsEncryptDisabled()))

	log.Fatal(p.Start())
}

func loadConfig() *Config {
	config := new(Config)

	defaultSslInsecure := envConfigBoolWithDefault("SSL_INSECURE", true)
	defaultCertPath := envConfigStringWithDefault("PROXY_CERT_PATH", "/proxy/certs")
	defaultDebug := envConfigIntWithDefault("DEBUG_LEVEL", 0)
	defaultKmsResourceName := envConfigStringWithDefault("GCP_KMS_RESOURCE_NAME", "")

	flag.BoolVar(&config.version, "version", false, "show go-gcsproxy version")
	flag.StringVar(&config.Addr, "port", ":9080", "proxy listen addr")
	flag.StringVar(&config.WebAddr, "web_port", ":9081", "web interface listen addr")
	flag.BoolVar(&config.SslInsecure, "ssl_insecure", defaultSslInsecure, "don't verify upstream server SSL/TLS certificates.")

	flag.StringVar(&config.CertPath, "cert_path", defaultCertPath, "path to cert. if 'mitmproxy-ca.pem' is not present here, it will be generated.")
	flag.IntVar(&config.Debug, "debug", defaultDebug, "debug level: 0 - ERROR, 1 - DEBUG, 2 - TRACE")
	flag.StringVar(&config.Dump, "dump", "", "filename to dump req/responses for debugging")
	flag.IntVar(&config.DumpLevel, "dump_level", 0, "dump level: 0 - header, 1 - header + body")
	flag.StringVar(&config.Upstream, "upstream", "", "upstream proxy")
	flag.StringVar(&config.KmsResourceName, "kms_resource_name", defaultKmsResourceName, "payload will be encrypted with this key stored in KMS. Must be in the format: projects/<project_id>/locations/<global|region>/keyRings/<key_ring>/cryptoKeys/<key>")

	flag.BoolVar(&config.UpstreamCert, "upstream_cert", false, "connect to upstream server to look up certificate details")
	flag.Parse()

	return config
}
func Usage() {
	flag.Usage()
	fmt.Println("\nEnvironment variables supported:")
	fmt.Println("  GCP_KMS_RESOURCE_NAME")
	fmt.Println("  PROXY_CERT_PATH")
	fmt.Println("  SSL_INSECURE")
	fmt.Println("  DEBUG_LEVEL")
}

func CheckKMS() error {
	var ctx = context.TODO()

	_, err := encryptBytes(ctx, config.KmsResourceName, []byte("Hello, World!"))
	return err
}
