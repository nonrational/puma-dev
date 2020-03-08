package dev

import (
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testPumaDevRequest PumaDevRequest

type PumaDevRequestCtx struct {
	URL            string
	Host           string
	PumaDevHost    string
	PumaDevAppName string
}

func (ctx *PumaDevRequestCtx) PumaDevReq() PumaDevRequest {
	var r io.Reader

	req, _ := http.NewRequest("GET", ctx.URL, r)
	req.Host = req.URL.Host

	if ctx.PumaDevAppName != "" {
		req.Header.Add("Puma-Dev-App-Name", ctx.PumaDevAppName)
	}

	if ctx.Host != "" {
		req.Header.Add("Host", ctx.Host)
		req.Host = ctx.Host
	}

	if ctx.PumaDevHost != "" {
		req.Header.Add("Puma-Dev-Host", ctx.PumaDevHost)
	}

	return PumaDevRequest{req}
}

func TestPumaDevRequest_AppName(t *testing.T) {

	testCases := map[PumaDevRequestCtx]string{
		PumaDevRequestCtx{URL: "http://qwerty.puma", PumaDevHost: "asdf.puma"}: "asdf",
		PumaDevRequestCtx{URL: "http://qwerty.puma", PumaDevAppName: "asdf"}:   "asdf",
		PumaDevRequestCtx{URL: "http://qwerty.puma", Host: "asdf.puma"}:        "asdf",
		PumaDevRequestCtx{URL: "http://qwerty.puma"}:                           "qwerty",

		PumaDevRequestCtx{URL: "https://127.0.0.1:443", PumaDevHost: "asdf.puma"}: "asdf",
		PumaDevRequestCtx{URL: "https://127.0.0.1:443", PumaDevAppName: "asdf"}:   "asdf",
		PumaDevRequestCtx{URL: "https://127.0.0.1:443", Host: "asdf.puma"}:        "asdf",
		PumaDevRequestCtx{URL: "https://my.app.with.puma"}:                        "my.app.with",

		PumaDevRequestCtx{URL: "http://127.0.0.1", Host: "proxy.io", PumaDevAppName: "asdf"}:   "asdf",
		PumaDevRequestCtx{URL: "http://127.0.0.1", Host: "proxy.io", PumaDevHost: "asdf.puma"}: "asdf",
	}

	for ctx, expectedAppName := range testCases {
		t.Run(ctx.URL, func(t *testing.T) {
			pdReq := ctx.PumaDevReq()
			assert.Equal(t, expectedAppName, pdReq.AppName())
		})
	}
}

func TestPumaDevRequest_canonicalAppNameFromHost(t *testing.T) {
	testCases := map[string]string{
		"app":                        "app",
		"app.test":                   "app",
		"app.dev:8080":               "app",
		"app.0.0.0.0.xip.io":         "app",
		"app.255.255.255.255.nip.io": "app",
		"app.loc.al":                 "app.loc",
		"app.0.0.xip.io":             "",
	}

	for host, expectedAppName := range testCases {
		t.Run(host, func(t *testing.T) {
			appName := testPumaDevRequest.canonicalAppNameFromHost(host)
			assert.Equal(t, appName, expectedAppName)
		})
	}
}
