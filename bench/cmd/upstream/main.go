package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	var (
		listen = flag.String("listen", ":18080", "listen address")
		path   = flag.String("path", "/bench", "benchmark path")
	)
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Upstream", "bench")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":        true,
			"path":      r.URL.Path,
			"timestamp": time.Now().UnixNano(),
		})
	}
	// Handle both the exact benchmark path and any forwarded path variants from proxy rewrites.
	mux.HandleFunc(*path, handler)
	mux.HandleFunc("/", handler)

	log.Printf("upstream listening on %s", *listen)
	log.Fatal(http.ListenAndServe(*listen, mux))
}
