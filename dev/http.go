package dev

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/bmizerany/pat"
	"github.com/puma/puma-dev/httpu"
	"github.com/puma/puma-dev/httputil"
)

type HTTPServer struct {
	Address    string
	TLSAddress string
	Pool       *AppPool
	Debug      bool
	Events     *Events

	mux       *pat.PatternServeMux
	transport *httpu.Transport
	proxy     *httputil.ReverseProxy
}

func (h *HTTPServer) Setup() {
	h.transport = &httpu.Transport{
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	h.Debug = true

	h.Pool.AppClosed = h.AppClosed

	h.proxy = &httputil.ReverseProxy{
		Proxy:         h.proxyReq,
		Transport:     h.transport,
		FlushInterval: 1 * time.Second,
		Debug:         h.Debug,
	}

	h.mux = pat.New()

	h.mux.Get("/status", http.HandlerFunc(h.status))
	h.mux.Get("/events", http.HandlerFunc(h.events))
}

func (h *HTTPServer) AppClosed(app *App) {
	// Whenever an app is closed, wipe out all idle conns. This
	// obviously closes down more than just this one apps connections
	// but that's ok.
	h.transport.CloseIdleConnections()
}

func (h *HTTPServer) findFirstApp(names []string) (*App, error) {
	var (
		app *App
		err error
	)

	for _, name := range names {
		app, err = h.Pool.App(name)
		if err != nil {
			if err == ErrUnknownApp {
				continue
			}

			return nil, err
		}
	}

	if app == nil {
		app, err = h.Pool.App("default")
		if err != nil {
			return nil, err
		}
	}

	err = app.WaitTilReady()
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (h *HTTPServer) proxyReq(w http.ResponseWriter, req *http.Request) error {
	pdReq := PumaDevRequest{req}

	app, err := h.findFirstApp(pdReq.AllAppNames())

	if err != nil {
		if err == ErrUnknownApp {
			h.Events.Add("unknown_app", "name", pdReq.AppName(), "host", req.Host)
		} else {
			h.Events.Add("lookup_error", "error", err.Error())
		}

		return err
	}

	if app.Public && req.URL.Path != "/" {
		safeURLPath := path.Clean(req.URL.Path)
		path := filepath.Join(app.dir, "public", safeURLPath)

		fi, err := os.Stat(path)
		if err == nil && !fi.IsDir() {
			if ofile, err := os.Open(path); err == nil {
				http.ServeContent(w, req, req.URL.Path, fi.ModTime(), io.ReadSeeker(ofile))
				return httputil.ErrHandled
			}
		}
	}

	req.URL.Scheme, req.URL.Host = app.Scheme, app.Address()
	return err
}

func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pdReq := PumaDevRequest{req}

	if h.Debug {
		fmt.Fprintf(os.Stderr, "%s: %s '%s' (host=%s)\n",
			time.Now().Format(time.RFC3339Nano),
			req.Method, req.URL.Path, req.Host)
	}

	if pdReq.AppName() == "puma-dev" {
		h.mux.ServeHTTP(w, req)
	} else {
		h.proxy.ServeHTTP(w, req)
	}
}

func (h *HTTPServer) status(w http.ResponseWriter, req *http.Request) {
	type appStatus struct {
		Scheme  string `json:"scheme"`
		Address string `json:"address"`
		Status  string `json:"status"`
		Log     string `json:"log"`
	}

	statuses := map[string]appStatus{}

	h.Pool.ForApps(func(a *App) {
		var status string

		switch a.Status() {
		case Dead:
			status = "dead"
		case Booting:
			status = "booting"
		case Running:
			status = "running"
		default:
			status = "unknown"
		}

		statuses[a.Name] = appStatus{
			Scheme:  a.Scheme,
			Address: a.Address(),
			Status:  status,
			Log:     a.Log(),
		}
	})

	json.NewEncoder(w).Encode(statuses)
}

func (h *HTTPServer) events(w http.ResponseWriter, req *http.Request) {
	h.Events.WriteTo(w)
}
