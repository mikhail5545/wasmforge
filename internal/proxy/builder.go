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

type (
	Builder interface {
		Director() *Director
		BuildRoute(targetURL, path string, transportCfg TransportConfig, middlewares ...middleware) error
		RebuildRouteMiddlewares(path string, middlewares ...middleware) error
		RemoveRoute(path string) error
	}

	builder struct {
		mu       sync.RWMutex
		routes   map[string]*internalRoute
		director *Director
	}

	internalRoute struct {
		handler http.Handler           // Fully assembled chain: WASM(s) -> Proxy, this will be called when request matches the route
		proxy   *httputil.ReverseProxy // Original reverse proxy instance without middleware applied, used for hot-swapping middlewares without re-parsing the target URL
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
		IdleConnTimeout       time.Duration
		TLSHandshakeTimeout   time.Duration
		ExpectContinueTimeout time.Duration
		MaxIdleCons           *int
		MaxIdleConsPerHost    *int
		MaxConsPerHost        *int
		ResponseHeaderTimeout *time.Duration
	}

	middleware func(http.Handler) http.Handler
)

func NewBuilder() Builder {
	return &builder{
		routes: make(map[string]*internalRoute),
	}
}

func (b *builder) Director() *Director {
	return b.director
}

func (b *builder) BuildRoute(targetURL, path string, transportCfg TransportConfig, middlewares ...middleware) error {
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
		req.URL.Path = strings.TrimPrefix(req.URL.Path, target.Path)
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
	b.director.AddRoute(path, final)
	return nil
}

func (b *builder) RebuildRouteMiddlewares(path string, middlewares ...middleware) error {
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
	b.director.AddRoute(path, final)
	return nil
}

func (b *builder) RemoveRoute(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.routes[path]; !exists {
		return fmt.Errorf("route not found: %s", path)
	}
	delete(b.routes, path)
	return b.director.RemoveRoute(path)
}

func newTransport(cfg TransportConfig) *http.Transport {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		IdleConnTimeout:       cfg.Timeout.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.Timeout.TLSHandshakeTimeout,
		ExpectContinueTimeout: cfg.Timeout.ExpectContinueTimeout,
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
