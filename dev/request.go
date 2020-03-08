package dev

import (
	"net"
	"net/http"
	"strings"
)

type PumaDevRequest struct {
	httpRequest *http.Request
}

func pruneSubdomain(name string) string {
	dot := strings.IndexByte(name, '.')
	if dot == -1 {
		return ""
	}

	return name[dot+1:]
}

func (r *PumaDevRequest) AllAppNames() []string {
	currAppName := r.AppName()

	allAppNames := []string{currAppName}

	for nextName := pruneSubdomain(currAppName); nextName != ""; {
		allAppNames = append(allAppNames, nextName)
		currAppName = nextName
	}

	return allAppNames
}

func (r *PumaDevRequest) AppName() string {
	if reqHeaderAppName := r.httpRequest.Header.Get("Puma-Dev-App-Name"); reqHeaderAppName != "" {
		return reqHeaderAppName
	}

	if reqHeaderHost := r.httpRequest.Header.Get("Puma-Dev-Host"); reqHeaderHost != "" {
		return r.canonicalAppNameFromHost(reqHeaderHost)
	}

	return r.canonicalAppNameFromHost(r.httpRequest.Host)
}

func (r *PumaDevRequest) canonicalAppNameFromHost(host string) string {
	colon := strings.LastIndexByte(host, ':')
	if colon != -1 {
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
	}

	if strings.HasSuffix(host, ".xip.io") || strings.HasSuffix(host, ".nip.io") {
		parts := strings.Split(host, ".")
		if len(parts) < 6 {
			return ""
		}

		name := strings.Join(parts[:len(parts)-6], ".")

		return name
	}

	dot := strings.LastIndexByte(host, '.')

	if dot == -1 {
		return host
	} else {
		return host[:dot]
	}
}
