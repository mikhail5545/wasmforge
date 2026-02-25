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

package wasm

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/mikhail5545/wasmforge/internal/proxy/reqctx"
	"github.com/mikhail5545/wasmforge/internal/util/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"go.uber.org/zap"
)

func setupLogger() (*zap.Logger, func(), error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		_ = logger.Sync()
	}
	return logger, cleanup, nil
}

func setupTest(t *testing.T) (*zap.Logger, wazero.Runtime, func(wazero.Runtime), context.Context) {
	logger, cleanup, err := setupLogger()
	if err != nil {
		t.Fatalf("Failed to set up logger: %v", err)
	}
	t.Cleanup(cleanup)

	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	closeRt := func(runtime wazero.Runtime) {
		if err := runtime.Close(ctx); err != nil {
			t.Errorf("Failed to close WASM runtime: %v", err)
		}
	}

	return logger, runtime, closeRt, ctx
}

func TestHostGetHeader(t *testing.T) {
	logger, runtime, closeRt, ctx := setupTest(t)
	defer closeRt(runtime)

	guestModuleBinary := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, // WASM binary header
		0x05, 0x03, 0x01, 0x00, 0x01, // Memory section: 1 page of memory (64KB)
		0x07, 0x0a, 0x01, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x02, 0x00, // Export memory as "memory"
	}

	compiledMod, err := runtime.CompileModule(ctx, guestModuleBinary)
	require.NoError(t, err, "Failed to compile guest module")

	for _, tt := range []struct {
		name        string
		headerKey   string
		headerVal   string
		searchKey   string
		bufPtr      uint32
		bufMaxLen   uint32
		wantLen     uint32
		wantValue   string
		expectFound bool
	}{
		{
			name:        "Found exact match",
			headerKey:   "Content-Type",
			headerVal:   "application/json",
			searchKey:   "Content-Type",
			bufPtr:      200,
			bufMaxLen:   50,
			wantLen:     uint32(len("application/json")),
			wantValue:   "application/json",
			expectFound: true,
		},
		{
			name:        "Found with truncation",
			headerKey:   "Content-Type",
			headerVal:   "application/json",
			searchKey:   "Content-Type",
			bufPtr:      300,
			bufMaxLen:   10,
			wantLen:     10,
			wantValue:   "applicatio",
			expectFound: true,
		},
		{
			name:        "Header not found",
			headerKey:   "Content-Type",
			headerVal:   "application/json",
			searchKey:   "Authorization",
			bufPtr:      300,
			bufMaxLen:   50,
			wantLen:     0xFFFFFFFF,
			wantValue:   "",
			expectFound: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/test", nil)
			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.headerVal)
			}
			ctx = reqctx.WithRequest(ctx, req)
			ctx = reqctx.WithLogger(ctx, logger)

			mod, err := runtime.InstantiateModule(ctx, compiledMod, wazero.NewModuleConfig().WithName(tt.name))
			require.NoError(t, err)
			defer func(mod api.Module) {
				if err := mod.Close(ctx); err != nil {
					t.Errorf("Failed to close module: %v", err)
				}
			}(mod)

			keyPtr := uint32(0)
			keyLen := uint32(len(tt.searchKey))
			ok := mod.Memory().Write(keyPtr, []byte(tt.searchKey))
			assert.True(t, ok, "Failed to write header key to WASM memory")

			resLen := hostGetHeader(ctx, mod, keyPtr, keyLen, tt.bufPtr, tt.bufMaxLen)
			if tt.expectFound {
				require.Equal(t, tt.wantLen, resLen, "Expected header value length to match")
				valBytes, ok := mod.Memory().Read(tt.bufPtr, resLen)
				require.True(t, ok, "Failed to read header value from WASM memory")
				require.Equal(t, tt.wantValue, string(valBytes), "Expected header value to match")
			} else {
				require.Equal(t, tt.wantLen, resLen, "Expected header not to be found")
			}
		})
	}
}

func TestHostSetHeader(t *testing.T) {
	logger, runtime, closeRt, ctx := setupTest(t)
	defer closeRt(runtime)

	guestModuleBinary := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, // WASM binary header
		0x05, 0x03, 0x01, 0x00, 0x01, // Memory section: 1 page of memory (64KB)
		0x07, 0x0a, 0x01, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x02, 0x00, // Export memory as "memory"
	}

	compiledMod, err := runtime.CompileModule(ctx, guestModuleBinary)
	require.NoError(t, err, "Failed to compile guest module")

	for _, tt := range []struct {
		name      string
		headerKey string
		headerVal string
		wroteLen  uint32
	}{
		{
			name:      "Set header successfully",
			headerKey: "X-Test-Header",
			headerVal: "TestValue",
			wroteLen:  uint32(len("TestValue")),
		},
		{
			name:      "Set header with empty value",
			headerKey: "X-Empty-Header",
			headerVal: "",
			wroteLen:  0,
		},
		{
			name:      "Set header with long value",
			headerKey: "X-Long-Header",
			headerVal: "This is a very long header value that exceeds typical lengths",
			wroteLen:  uint32(len("This is a very long header value that exceeds typical lengths")),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/test", nil)
			ctx = reqctx.WithRequest(ctx, req)
			ctx = reqctx.WithLogger(ctx, logger)

			mod, err := runtime.InstantiateModule(ctx, compiledMod, wazero.NewModuleConfig().WithName(tt.name))
			require.NoError(t, err)
			defer func(mod api.Module) {
				if err := mod.Close(ctx); err != nil {
					t.Errorf("Failed to close module: %v", err)
				}
			}(mod)

			keyPtr := uint32(0)
			keyLen := uint32(len(tt.headerKey))
			valPtr := uint32(100)
			valLen := uint32(len(tt.headerVal))

			ok := mod.Memory().Write(keyPtr, []byte(tt.headerKey))
			assert.True(t, ok, "Failed to write header key to WASM memory")
			ok = mod.Memory().Write(valPtr, []byte(tt.headerVal))
			assert.True(t, ok, "Failed to write header value to WASM memory")

			hostSetHeader(ctx, mod, keyPtr, keyLen, valPtr, valLen)

			gotVal := req.Header.Get(tt.headerKey)
			require.Equal(t, tt.headerVal, gotVal, "Expected header value to match what was set")
		})
	}
}

func TestHostGetMethod(t *testing.T) {
	logger, runtime, closeRt, ctx := setupTest(t)
	defer closeRt(runtime)

	guestModuleBinary := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, // WASM binary header
		0x05, 0x03, 0x01, 0x00, 0x01, // Memory section: 1 page of memory (64KB)
		0x07, 0x0a, 0x01, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x02, 0x00, // Export memory as "memory"
	}

	compiledMod, err := runtime.CompileModule(ctx, guestModuleBinary)
	require.NoError(t, err, "Failed to compile guest module")

	for _, tt := range []struct {
		name       string
		method     string
		bufPtr     uint32
		bufMaxLen  uint32
		wantLen    uint32
		wantMethod string
	}{
		{
			name:       "Get GET method",
			method:     "GET",
			bufPtr:     200,
			bufMaxLen:  10,
			wantLen:    uint32(len("GET")),
			wantMethod: "GET",
		},
		{
			name:       "Get POST method",
			method:     "POST",
			bufPtr:     300,
			bufMaxLen:  10,
			wantLen:    uint32(len("POST")),
			wantMethod: "POST",
		},
		{
			name:       "Truncate long method",
			method:     "LONGMETHODNAME",
			bufPtr:     400,
			bufMaxLen:  4,
			wantLen:    4,
			wantMethod: "LONG",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "http://example.com/test", nil)
			ctx = reqctx.WithRequest(ctx, req)
			ctx = reqctx.WithLogger(ctx, logger)

			mod, err := runtime.InstantiateModule(ctx, compiledMod, wazero.NewModuleConfig().WithName(tt.name))
			require.NoError(t, err)
			defer func(mod api.Module) {
				if err := mod.Close(ctx); err != nil {
					t.Errorf("Failed to close module: %v", err)
				}
			}(mod)

			resLen := hostGetMethod(ctx, mod, tt.bufPtr, tt.bufMaxLen)
			require.Equal(t, tt.wantLen, resLen, "Expected method length to match")
			methodBytes, ok := mod.Memory().Read(tt.bufPtr, resLen)
			require.True(t, ok, "Failed to read method from WASM memory")
			require.Equal(t, tt.wantMethod, string(methodBytes), "Expected method to match")
		})
	}
}

func TestHostGetPath(t *testing.T) {
	logger, runtime, closeRt, ctx := setupTest(t)
	defer closeRt(runtime)

	guestModuleBinary := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, // WASM binary header
		0x05, 0x03, 0x01, 0x00, 0x01, // Memory section: 1 page of memory (64KB)
		0x07, 0x0a, 0x01, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x02, 0x00, // Export memory as "memory"
	}

	compiledMod, err := runtime.CompileModule(ctx, guestModuleBinary)
	require.NoError(t, err, "Failed to compile guest module")

	for _, tt := range []struct {
		name      string
		path      string
		bufPtr    uint32
		bufMaxLen uint32
		wantLen   uint32
		wantPath  string
	}{
		{
			name:      "Get simple path",
			path:      "/test",
			bufPtr:    200,
			bufMaxLen: 50,
			wantLen:   uint32(len("/test")),
			wantPath:  "/test",
		},
		{
			name:      "Get path from req with query params",
			path:      "/test?param=value&param2=value2",
			bufPtr:    300,
			bufMaxLen: 50,
			wantLen:   uint32(len("/test")),
			wantPath:  "/test",
		},
		{
			name:      "Truncate long path",
			path:      "/this/is/a/very/long/path/that/exceeds/the/buffer/size",
			bufPtr:    400,
			bufMaxLen: 20,
			wantLen:   20,
			wantPath:  "/this/is/a/very/long",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com"+tt.path, nil)
			ctx = reqctx.WithRequest(ctx, req)
			ctx = reqctx.WithLogger(ctx, logger)

			mod, err := runtime.InstantiateModule(ctx, compiledMod, wazero.NewModuleConfig().WithName(tt.name))
			require.NoError(t, err)
			defer func(mod api.Module) {
				if err := mod.Close(ctx); err != nil {
					t.Errorf("Failed to close module: %v", err)
				}
			}(mod)

			resLen := hostGetPath(ctx, mod, tt.bufPtr, tt.bufMaxLen)
			require.Equal(t, tt.wantLen, resLen, "Expected path length to match")
			pathBytes, ok := mod.Memory().Read(tt.bufPtr, resLen)
			require.True(t, ok, "Failed to read path from WASM memory")
			require.Equal(t, tt.wantPath, string(pathBytes), "Expected path to match")
		})
	}
}

func TestHostGetQueryParam(t *testing.T) {
	logger, runtime, closeRt, ctx := setupTest(t)
	defer closeRt(runtime)

	guestModuleBinary := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, // WASM binary header
		0x05, 0x03, 0x01, 0x00, 0x01, // Memory section: 1 page of memory (64KB)
		0x07, 0x0a, 0x01, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x02, 0x00, // Export memory as "memory"
	}

	compiledMod, err := runtime.CompileModule(ctx, guestModuleBinary)
	require.NoError(t, err, "Failed to compile guest module")

	for _, tt := range []struct {
		name      string
		url       string
		key       string
		value     string
		bufPtr    uint32
		bufMaxLen uint32
		wantLen   uint32
		wantValue string
	}{
		{
			name:      "Get existing query param",
			url:       "http://example.com/test?param=value",
			key:       "param",
			value:     "value",
			bufPtr:    200,
			bufMaxLen: 50,
			wantLen:   uint32(len("value")),
			wantValue: "value",
		},
		{
			name:      "Get non-existing query param",
			url:       "http://example.com/test?param=value",
			key:       "nonexistent",
			value:     "",
			bufPtr:    300,
			bufMaxLen: 50,
			wantLen:   0xFFFFFFFF,
			wantValue: "",
		},
		{
			name:      "Truncate long query param value",
			url:       "http://example.com/test?param=ThisIsAVeryLongValueThatExceedsTheBufferSize",
			key:       "param",
			value:     "ThisIsAVeryLongValueThatExceedsTheBufferSize",
			bufPtr:    400,
			bufMaxLen: 10,
			wantLen:   10,
			wantValue: "ThisIsAVer",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			ctx = reqctx.WithRequest(ctx, req)
			ctx = reqctx.WithLogger(ctx, logger)

			mod, err := runtime.InstantiateModule(ctx, compiledMod, wazero.NewModuleConfig().WithName(tt.name))
			require.NoError(t, err)
			defer func(mod api.Module) {
				if err := mod.Close(ctx); err != nil {
					t.Errorf("Failed to close module: %v", err)
				}
			}(mod)

			keyPtr := uint32(0)
			keyLen := uint32(len(tt.key))
			ok := mod.Memory().Write(keyPtr, []byte(tt.key))
			assert.True(t, ok, "Failed to write query param key to WASM memory")

			resLen := hostGetQueryParam(ctx, mod, keyPtr, keyLen, tt.bufPtr, tt.bufMaxLen)
			if tt.wantValue != "" {
				require.Equal(t, tt.wantLen, resLen, "Expected query param value length to match")
				valBytes, ok := mod.Memory().Read(tt.bufPtr, resLen)
				require.True(t, ok, "Failed to read query param value from WASM memory")
				require.Equal(t, tt.wantValue, string(valBytes), "Expected query param value to match")
			} else {
				require.Equal(t, tt.wantLen, resLen, "Expected query param not to be found")
			}
		})
	}
}

func TestHostGetRawQuery(t *testing.T) {
	logger, runtime, closeRt, ctx := setupTest(t)
	defer closeRt(runtime)

	guestModuleBinary := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, // WASM binary header
		0x05, 0x03, 0x01, 0x00, 0x01, // Memory section: 1 page of memory (64KB)
		0x07, 0x0a, 0x01, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x02, 0x00, // Export memory as "memory"
	}

	compiledMod, err := runtime.CompileModule(ctx, guestModuleBinary)
	require.NoError(t, err, "Failed to compile guest module")

	for _, tt := range []struct {
		name      string
		url       string
		bufPtr    uint32
		bufMaxLen uint32
		wantLen   uint32
		wantQuery string
	}{
		{
			name:      "Get raw query string",
			url:       "http://example.com/test?param=value&param2=value2",
			bufPtr:    200,
			bufMaxLen: 50,
			wantLen:   uint32(len("param=value&param2=value2")),
			wantQuery: "param=value&param2=value2",
		},
		{
			name:      "Truncate long raw query string",
			url:       "http://example.com/test?param=ThisIsAVeryLongValueThatExceedsTheBufferSize",
			bufPtr:    300,
			bufMaxLen: 10,
			wantLen:   10,
			wantQuery: "param=This",
		},
		{
			name:      "No query string",
			url:       "http://example.com/test",
			bufPtr:    400,
			bufMaxLen: 50,
			wantLen:   0,
			wantQuery: "",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			ctx = reqctx.WithRequest(ctx, req)
			ctx = reqctx.WithLogger(ctx, logger)

			mod, err := runtime.InstantiateModule(ctx, compiledMod, wazero.NewModuleConfig().WithName(tt.name))
			require.NoError(t, err)
			defer func(mod api.Module) {
				if err := mod.Close(ctx); err != nil {
					t.Errorf("Failed to close module: %v", err)
				}
			}(mod)

			resLen := hostGetRawQuery(ctx, mod, tt.bufPtr, tt.bufMaxLen)
			require.Equal(t, tt.wantLen, resLen, "Expected raw query string length to match")
			queryBytes, ok := mod.Memory().Read(tt.bufPtr, resLen)
			require.True(t, ok, "Failed to read raw query string from WASM memory")
			require.Equal(t, tt.wantQuery, string(queryBytes), "Expected raw query string to match")
		})
	}
}

func TestHostSendResponse(t *testing.T) {
	logger, runtime, closeRt, ctx := setupTest(t)
	defer closeRt(runtime)

	guestModuleBinary := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, // WASM binary header
		0x05, 0x03, 0x01, 0x00, 0x01, // Memory section: 1 page of memory (64KB)
		0x07, 0x0a, 0x01, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x02, 0x00, // Export memory as "memory"
	}

	compiledMod, err := runtime.CompileModule(ctx, guestModuleBinary)
	require.NoError(t, err, "Failed to compile guest module")

	for _, tt := range []struct {
		name         string
		sendBody     bool
		body         string
		statusCode   uint32
		expectedCode int
	}{
		{
			name:         "Send response with body",
			sendBody:     true,
			body:         "Hello, World!",
			statusCode:   403,
			expectedCode: 403,
		},
		{
			name:         "Send response without body",
			sendBody:     false,
			body:         "",
			statusCode:   400,
			expectedCode: 400,
		},
		{
			name:         "Send response with long body",
			sendBody:     true,
			body:         "This is a very long response body that exceeds typical lengths and is used to test the handling of large response bodies in the hostSendResponse function.",
			statusCode:   422,
			expectedCode: 422,
		},
		{
			name:         "Send response with 200 OK status",
			sendBody:     true,
			body:         "This response has a 200 OK status code",
			statusCode:   200,
			expectedCode: 200,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/test", nil)
			ctx = reqctx.WithRequest(ctx, req)
			ctx = reqctx.WithLogger(ctx, logger)
			state := reqctx.RequestState{Interrupted: false}
			ctx = reqctx.WithRequestState(ctx, &state)

			mod, err := runtime.InstantiateModule(ctx, compiledMod, wazero.NewModuleConfig().WithName(tt.name))
			require.NoError(t, err)
			defer func(mod api.Module) {
				if err := mod.Close(ctx); err != nil {
					t.Errorf("Failed to close module: %v", err)
				}
			}(mod)

			bodyPtr := uint32(100)
			bodyLen := uint32(len(tt.body))
			if tt.sendBody {
				ok := mod.Memory().Write(bodyPtr, []byte(tt.body))
				assert.True(t, ok, "Failed to write response body to WASM memory")
			}

			hostSendResponse(ctx, mod, tt.statusCode, bodyPtr, bodyLen)

			assert.True(t, state.Interrupted, "Expected request to be marked as interrupted")
			assert.Equal(t, tt.expectedCode, state.StatusCode, "Expected response status code to match what was sent (with 200 treated as 500)")
			assert.Equal(t, tt.body, string(state.Body), "Expected context state response body to match what was sent")
			assert.True(t, state.Interrupted)
		})
	}
}

func TestHostGetJSONConfig(t *testing.T) {
	logger, runtime, closeRt, ctx := setupTest(t)
	defer closeRt(runtime)

	guestModuleBinary := []byte{
		0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, // WASM binary header
		0x05, 0x03, 0x01, 0x00, 0x01, // Memory section: 1 page of memory (64KB)
		0x07, 0x0a, 0x01, 0x06, 0x6d, 0x65, 0x6d, 0x6f, 0x72, 0x79, 0x02, 0x00, // Export memory as "memory"
	}

	compiledMod, err := runtime.CompileModule(ctx, guestModuleBinary)
	require.NoError(t, err, "Failed to compile guest module")

	for _, tt := range []struct {
		name      string
		config    *string
		bufMaxLen uint32
		wantLen   uint32
		wantJSON  string
	}{
		{
			name:      "Get existing JSON config",
			config:    memory.Ptr(`{"key": "value", "number": 42}`),
			wantLen:   uint32(len(`{"key": "value", "number": 42}`)),
			bufMaxLen: 50,
			wantJSON:  `{"key": "value", "number": 42}`,
		},
		{
			name:      "Get empty JSON config",
			config:    memory.Ptr(`{}`),
			wantLen:   uint32(len(`{}`)),
			bufMaxLen: 10,
			wantJSON:  `{}`,
		},
		{
			name:      "Truncate long JSON config",
			config:    memory.Ptr(`{"key": "This is a very long value that exceeds the buffer size"}`),
			wantLen:   20,
			bufMaxLen: 20,
			wantJSON:  `{"key": "This is a v`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx = reqctx.WithJSONConfig(ctx, tt.config)
			ctx = reqctx.WithLogger(ctx, logger)

			mod, err := runtime.InstantiateModule(ctx, compiledMod, wazero.NewModuleConfig().WithName(tt.name))
			require.NoError(t, err)
			defer func(mod api.Module) {
				if err := mod.Close(ctx); err != nil {
					t.Errorf("Failed to close module: %v", err)
				}
			}(mod)

			bufPtr := uint32(100)
			resLen := hostGetJSONConfig(ctx, mod, bufPtr, tt.bufMaxLen)
			require.Equal(t, tt.wantLen, resLen, "Expected JSON config length to match")
			jsonBytes, ok := mod.Memory().Read(bufPtr, resLen)
			require.True(t, ok, "Failed to read JSON config from WASM memory")
			require.Equal(t, tt.wantJSON, string(jsonBytes), "Expected JSON config to match what was set in context")
		})
	}
}
