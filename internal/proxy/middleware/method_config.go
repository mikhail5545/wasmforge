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

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	methodmodel "github.com/mikhail5545/wasmforge/internal/models/route/method"
	"golang.org/x/time/rate"
)

type methodRuntimeConfigKey struct{}

type MethodRuntimeConfig struct {
	Method                 string
	MaxRequestPayloadBytes *int64
	RequestTimeoutMs       *int
	ResponseTimeoutMs      *int
	RateLimiter            *rate.Limiter
	RequireAuthentication  bool
	AllowedAuthSchemes     map[string]struct{}
	Metadata               map[string]any
}

func MethodRuntimeConfigFromContext(ctx context.Context) (*MethodRuntimeConfig, bool) {
	cfg, ok := ctx.Value(methodRuntimeConfigKey{}).(*MethodRuntimeConfig)
	return cfg, ok
}

// NewMethodConfigMiddleware enforces per-method route settings and exposes
// method metadata to downstream auth/plugin middleware through request context.
func NewMethodConfigMiddleware(configMap map[string]methodmodel.RouteMethod) func(http.Handler) http.Handler {
	methodConfigs := make(map[string]*MethodRuntimeConfig, len(configMap))
	for m, cfg := range configMap {
		method := strings.ToUpper(strings.TrimSpace(m))
		if method == "" {
			method = strings.ToUpper(strings.TrimSpace(cfg.Method))
		}
		if method == "" {
			continue
		}

		mc := &MethodRuntimeConfig{
			Method:                 method,
			MaxRequestPayloadBytes: cfg.MaxRequestPayloadBytes,
			ResponseTimeoutMs:      positiveIntPtr(cfg.ResponseTimeoutMs),
			RequestTimeoutMs:       positiveIntPtr(cfg.RequestTimeoutMs),
			RequireAuthentication:  cfg.RequireAuthentication,
			AllowedAuthSchemes:     parseAllowedAuthSchemes(cfg.AllowedAuthSchemes),
			Metadata:               parseMetadata(cfg.Metadata),
		}
		if cfg.RateLimitPerMinute != nil && *cfg.RateLimitPerMinute > 0 {
			rps := float64(*cfg.RateLimitPerMinute) / 60.0
			mc.RateLimiter = rate.NewLimiter(rate.Limit(rps), *cfg.RateLimitPerMinute)
		}
		methodConfigs[method] = mc
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cfg, ok := methodConfigs[strings.ToUpper(r.Method)]
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), methodRuntimeConfigKey{}, cfg.cloneForRequest())
			r = r.WithContext(ctx)

			if cfg.MaxRequestPayloadBytes != nil && *cfg.MaxRequestPayloadBytes > 0 {
				r.Body = http.MaxBytesReader(w, r.Body, *cfg.MaxRequestPayloadBytes)
			}

			if cfg.RateLimiter != nil && !cfg.RateLimiter.Allow() {
				w.Header().Set("Retry-After", "60")
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			if cfg.RequireAuthentication && !isAuthorizationAllowed(r.Header.Get("Authorization"), cfg.AllowedAuthSchemes) {
				setAuthenticateHeader(w, cfg.AllowedAuthSchemes)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			if cfg.RequestTimeoutMs != nil {
				timeout := time.Duration(*cfg.RequestTimeoutMs) * time.Millisecond
				timeoutCtx, cancel := context.WithTimeout(r.Context(), timeout)
				defer cancel()
				r = r.WithContext(timeoutCtx)
			}

			if cfg.ResponseTimeoutMs != nil {
				serveWithResponseTimeout(w, r, next, time.Duration(*cfg.ResponseTimeoutMs)*time.Millisecond)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (c *MethodRuntimeConfig) cloneForRequest() *MethodRuntimeConfig {
	out := *c
	if c.AllowedAuthSchemes != nil {
		out.AllowedAuthSchemes = make(map[string]struct{}, len(c.AllowedAuthSchemes))
		for scheme := range c.AllowedAuthSchemes {
			out.AllowedAuthSchemes[scheme] = struct{}{}
		}
	}
	if c.Metadata != nil {
		out.Metadata = make(map[string]any, len(c.Metadata))
		for key, value := range c.Metadata {
			out.Metadata[key] = value
		}
	}
	return &out
}

func positiveIntPtr(value *int) *int {
	if value == nil || *value <= 0 {
		return nil
	}
	out := *value
	return &out
}

func parseAllowedAuthSchemes(raw string) map[string]struct{} {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var schemes []string
	if err := json.Unmarshal([]byte(raw), &schemes); err != nil {
		return nil
	}
	out := make(map[string]struct{}, len(schemes))
	for _, scheme := range schemes {
		normalized := strings.ToLower(strings.TrimSpace(scheme))
		if normalized != "" {
			out[normalized] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseMetadata(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var metadata map[string]any
	if err := json.Unmarshal([]byte(raw), &metadata); err != nil {
		return nil
	}
	return metadata
}

func isAuthorizationAllowed(header string, allowedSchemes map[string]struct{}) bool {
	header = strings.TrimSpace(header)
	if header == "" {
		return false
	}
	scheme, _, hasToken := strings.Cut(header, " ")
	if strings.TrimSpace(scheme) == "" || !hasToken || strings.TrimSpace(header[len(scheme):]) == "" {
		return false
	}
	if len(allowedSchemes) == 0 {
		return true
	}
	_, ok := allowedSchemes[strings.ToLower(scheme)]
	return ok
}

func setAuthenticateHeader(w http.ResponseWriter, allowedSchemes map[string]struct{}) {
	if len(allowedSchemes) == 0 {
		w.Header().Set("WWW-Authenticate", "Bearer")
		return
	}
	values := make([]string, 0, len(allowedSchemes))
	for scheme := range allowedSchemes {
		values = append(values, scheme)
	}
	w.Header().Set("WWW-Authenticate", strings.Join(values, ", "))
}

func serveWithResponseTimeout(w http.ResponseWriter, r *http.Request, next http.Handler, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	buffer := newBufferedResponseWriter()
	done := make(chan any, 1)
	go func() {
		defer func() {
			done <- recover()
		}()
		next.ServeHTTP(buffer, r.WithContext(ctx))
	}()

	select {
	case panicValue := <-done:
		if panicValue != nil {
			panic(panicValue)
		}
		buffer.flushTo(w)
	case <-ctx.Done():
		http.Error(w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
	}
}

type bufferedResponseWriter struct {
	mu          sync.Mutex
	header      http.Header
	body        bytes.Buffer
	status      int
	wroteHeader bool
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{
		header: make(http.Header),
		status: http.StatusOK,
	}
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.wroteHeader {
		return
	}
	w.status = statusCode
	w.wroteHeader = true
}

func (w *bufferedResponseWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.wroteHeader {
		w.status = http.StatusOK
		w.wroteHeader = true
	}
	return w.body.Write(data)
}

func (w *bufferedResponseWriter) flushTo(dst http.ResponseWriter) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for key, values := range w.header {
		for _, value := range values {
			dst.Header().Add(key, value)
		}
	}
	dst.WriteHeader(w.status)
	_, _ = dst.Write(w.body.Bytes())
}
