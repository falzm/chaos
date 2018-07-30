package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/falzm/chaos"
)

const defaultBindAddr = "127.0.0.1:8000"

var (
	flagURL                string
	flagBindAddr           string
	flagControllerBindAddr string
)

func init() {
	flag.StringVar(&flagURL, "url", "", "URL to upstream target")
	flag.StringVar(&flagBindAddr, "bind-addr", defaultBindAddr, "network address:port to bind proxy to")
	flag.StringVar(&flagControllerBindAddr, "controller-bind-addr", chaos.DefaultBindAddr,
		"network endpoint to bind chaos controller to")
	flag.Parse()
}

func main() {
	url, err := url.Parse(flagURL)
	if err != nil {
		log.Fatalf("invalid upstream URL: %s", err)
	}

	chaos, err := chaos.NewChaos(flagControllerBindAddr)
	if err != nil {
		log.Fatalf("unable to initialize chaos controller: %s", err)
	}

	if err := http.ListenAndServe(flagBindAddr,
		chaos.Handler(httputil.NewSingleHostReverseProxy(url).ServeHTTP)); err != nil {
		log.Fatalf("unable to initialize reverse proxy: %s", err)
	}
}
