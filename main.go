package main

import (
	"flag"
	"fmt"
	rawLog "log"
	"net/http"
	"os"

	"github.com/lqqyt2423/go-mitmproxy/addon"
	//"github.com/lqqyt2423/go-mitmproxy/internal/helper"
	"github.com/lqqyt2423/go-mitmproxy/proxy"
	"github.com/lqqyt2423/go-mitmproxy/web"
	log "github.com/sirupsen/logrus"
)

// makefile will turn this into a version
var Version = "undefined"

type Config struct {
	version bool // show version

	Addr        string   // proxy listen addr
	WebAddr     string   // web interface listen addr
	SslInsecure bool     // not verify upstream server SSL/TLS certificates.
	IgnoreHosts []string // a list of ignore hosts
	AllowHosts  []string // a list of allow hosts
	CertPath    string   // path of generate cert files
	Debug       int      // debug mode: 1 - print debug log, 2 - show debug from
	Dump        string   // dump filename
	DumpLevel   int      // dump level: 0 - header, 1 - header + body

	Upstream     string // upstream proxy
	UpstreamCert bool   // Connect to upstream server to look up certificate details. Default: True
}

func main() {
	config := loadConfig()

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

	opts := &proxy.Options{
		Debug:             config.Debug,
		Addr:              config.Addr,
		StreamLargeBodies: 1024 * 1024 * 5,
		SslInsecure:       config.SslInsecure,
		CaRootPath:        config.CertPath,
		Upstream:          config.Upstream,
	}

	p, err := proxy.NewProxy(opts)
	if err != nil {
		log.Fatal(err)
	}

	if config.version {
		fmt.Println("go-gcsproxy: " + Version)
		os.Exit(0)
	}

	log.Infof("go-gcsproxy version %v\n", Version)

	if len(config.IgnoreHosts) > 0 {
		p.SetShouldInterceptRule(func(req *http.Request) bool {
			return false
			//return !helper.MatchHost(req.Host, config.IgnoreHosts)
		})
	}
	if len(config.AllowHosts) > 0 {
		p.SetShouldInterceptRule(func(req *http.Request) bool {
			return true // helper.MatchHost(req.Host, config.AllowHosts)
		})
	}

	if !config.UpstreamCert {
		p.AddAddon(proxy.NewUpstreamCertAddon(false))
		log.Infoln("UpstreamCert config false")
	}

	p.AddAddon(&proxy.LogAddon{})
	p.AddAddon(web.NewWebAddon(config.WebAddr))

	p.AddAddon(&EncryptGcsPayload{})
	p.AddAddon(&DecryptGcsPayload{})

	if config.Dump != "" {
		dumper := addon.NewDumperWithFilename(config.Dump, config.DumpLevel)
		p.AddAddon(dumper)
	}

	log.Fatal(p.Start())
}

func loadConfig() *Config {
	config := new(Config)

	flag.BoolVar(&config.version, "version", false, "show go-gcsproxy version")
	flag.StringVar(&config.Addr, "addr", ":9080", "proxy listen addr")
	flag.StringVar(&config.WebAddr, "web_addr", ":9081", "web interface listen addr")
	flag.BoolVar(&config.SslInsecure, "ssl_insecure", true, "not verify upstream server SSL/TLS certificates.")
	flag.Var((*arrayValue)(&config.IgnoreHosts), "ignore_hosts", "a list of ignore hosts")
	flag.Var((*arrayValue)(&config.AllowHosts), "allow_hosts", "a list of allow hosts")
	flag.StringVar(&config.CertPath, "cert_path", "/Users/byronwhitlock/certs", "path of generate cert files")
	flag.IntVar(&config.Debug, "debug", 2, "debug mode: 1 - print debug log, 2 - show debug from")
	flag.StringVar(&config.Dump, "dump", "", "dump filename")
	flag.IntVar(&config.DumpLevel, "dump_level", 0, "dump level: 0 - header, 1 - header + body")
	flag.StringVar(&config.Upstream, "upstream", "", "upstream proxy")
	flag.BoolVar(&config.UpstreamCert, "upstream_cert", false, "connect to upstream server to look up certificate details")
	flag.Parse()

	return config
}

type arrayValue []string

func (a *arrayValue) String() string {
	return fmt.Sprint(*a)
}

func (a *arrayValue) Set(value string) error {
	*a = append(*a, value)
	return nil
}
