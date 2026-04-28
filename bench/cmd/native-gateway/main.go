package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	var (
		listen          = flag.String("listen", ":19100", "listen address")
		upstreamURL     = flag.String("upstream", "http://127.0.0.1:18080", "upstream base URL")
		pathPrefix      = flag.String("path", "/bench", "route path prefix")
		authHeaderName  = flag.String("auth-header-name", "X-API-Key", "auth header name")
		authHeaderValue = flag.String("auth-header-value", "bench-secret", "auth header value")
	)
	flag.Parse()

	target, err := url.Parse(*upstreamURL)
	if err != nil {
		log.Fatalf("invalid upstream URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path + strings.TrimPrefix(req.URL.Path, *pathPrefix)
		if target.RawQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = target.RawQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = target.RawQuery + "&" + req.URL.RawQuery
		}
		originalDirector(req)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.Handle(*pathPrefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(*authHeaderName) != *authHeaderValue {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		proxy.ServeHTTP(w, r)
	}))

	log.Printf("native gateway listening on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, mux))
}
