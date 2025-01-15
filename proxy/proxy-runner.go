/*
Copyright 2025 Google.

This software is provided as-is, without warranty or representation for any use or purpose.
*/
package proxy

import (
	cfg "github.com/byronwhitlock-google/go-gcsproxy/config"

	"github.com/byronwhitlock-google/go-mitmproxy/addon"
	"github.com/byronwhitlock-google/go-mitmproxy/proxy"
	"github.com/byronwhitlock-google/go-mitmproxy/web"
	log "github.com/sirupsen/logrus"
)

type ProxyRunner struct {
	proxy  *proxy.Proxy
	config *cfg.Config
}

func NewProxyRunner(config *cfg.Config) *ProxyRunner {
	return &ProxyRunner{config: config}
}

func (r *ProxyRunner) Start() error {
	opts := &proxy.Options{
		Debug:             r.config.Debug,
		Addr:              r.config.Addr,
		StreamLargeBodies: 1024 * 1024 * 1024 * 1024 * 10, // TODO: we need to implement streaming intercept functions set to 10TB for now!
		SslInsecure:       r.config.SslInsecure,
		CaRootPath:        r.config.CertPath,
		Upstream:          r.config.Upstream,
	}

	p, err := proxy.NewProxy(opts)
	if err != nil {
		log.Fatal(err)
	}

	if !r.config.UpstreamCert {
		p.AddAddon(proxy.NewUpstreamCertAddon(false))
	}

	p.AddAddon(&proxy.LogAddon{})
	p.AddAddon(web.NewWebAddon(r.config.WebAddr))

	p.AddAddon(&EncryptGcsPayload{})
	p.AddAddon(&DecryptGcsPayload{})
	p.AddAddon(&GetReqHeader{})

	if r.config.Dump != "" {
		dumper := addon.NewDumperWithFilename(r.config.Dump, r.config.DumpLevel)
		p.AddAddon(dumper)
	}

	return p.Start()
}
