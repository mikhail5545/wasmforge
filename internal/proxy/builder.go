/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"
)

//go:generate mockgen -destination=../mocks/proxy/builder.go -package=proxy . Builder

type (
	Builder interface {
		Director() *Director
		BuildRoute(targetURL, path string, transportCfg TransportConfig, middlewares ...func(http.Handler) http.Handler) error
		RebuildRouteMiddlewares(path string, middlewares ...func(http.Handler) http.Handler) error
		RemoveRoute(path string) error
	}

	builder struct {
		mu       sync.RWMutex
		routes   map[string]*internalRoute
		director *Director
	}

	internalRoute struct {
		handler http.Handler // Fully assembled chain: WASM middleware(s) -> Proxy, this will be called when request matches the route
		// Original reverse httputil.ReverseProxy instance without middleware applied, used for hot-swapping middlewares without re-initializing the instance itself
		proxy *httputil.ReverseProxy
	}

	TransportConfig struct {
		Conn    ConsConfig
		Timeout TimeoutConfig
	}

	ConsConfig struct {
		MaxIdleCons        *int
		MaxIdleConsPerHost *int
		MaxConsPerHost     *int
	}

	TimeoutConfig struct {
		IdleConnTimeout       int
		TLSHandshakeTimeout   int
		ExpectContinueTimeout int
		ResponseHeaderTimeout *int
	}
)

func NewBuilder() Builder {
	d := &Director{
		routes: make(map[string]http.Handler),
	}
	d.mux.Store(http.NewServeMux())
	return &builder{
		director: d,
		routes:   make(map[string]*internalRoute),
	}
}

func (b *builder) Director() *Director {
	return b.director
}

func (b *builder) BuildRoute(targetURL, path string, transportCfg TransportConfig, middlewares ...func(http.Handler) http.Handler) error {
	target, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = newTransport(transportCfg)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path + strings.TrimPrefix(req.URL.Path, path)
		if target.RawQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = target.RawQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = target.RawQuery + "&" + req.URL.RawQuery
		}
		originalDirector(req)
	}

	var final http.Handler = proxy
	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](final)
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	b.routes[path] = &internalRoute{
		handler: final,
		proxy:   proxy,
	}
	b.director.addRoute(path, final)
	return nil
}

func (b *builder) RebuildRouteMiddlewares(path string, middlewares ...func(http.Handler) http.Handler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	route, exists := b.routes[path]
	if !exists {
		return fmt.Errorf("route not found: %s", path)
	}

	var final http.Handler = route.proxy
	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](final)
	}
	route.handler = final
	b.director.addRoute(path, final)
	return nil
}

func (b *builder) RemoveRoute(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.routes[path]; !exists {
		return fmt.Errorf("route not found: %s", path)
	}
	delete(b.routes, path)
	return b.director.removeRoute(path)
}

func newTransport(cfg TransportConfig) *http.Transport {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		IdleConnTimeout:       time.Duration(cfg.Timeout.IdleConnTimeout) * time.Second,
		TLSHandshakeTimeout:   time.Duration(cfg.Timeout.TLSHandshakeTimeout) * time.Second,
		ExpectContinueTimeout: time.Duration(cfg.Timeout.ExpectContinueTimeout) * time.Second,
	}
	if cfg.Timeout.ResponseHeaderTimeout != nil {
		transport.ResponseHeaderTimeout = time.Duration(*cfg.Timeout.ResponseHeaderTimeout) * time.Second
	}
	switch {
	case cfg.Conn.MaxIdleCons != nil:
		transport.MaxIdleConns = *cfg.Conn.MaxIdleCons
	case cfg.Conn.MaxIdleConsPerHost != nil:
		transport.MaxIdleConnsPerHost = *cfg.Conn.MaxIdleConsPerHost
	case cfg.Conn.MaxConsPerHost != nil:
		transport.MaxConnsPerHost = *cfg.Conn.MaxConsPerHost
	}
	return transport
}
