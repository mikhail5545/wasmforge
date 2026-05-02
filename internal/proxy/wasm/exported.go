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
	"encoding/json"
	"fmt"

	"github.com/mikhail5545/wasmforge/internal/proxy/reqctx"
	"github.com/tetratelabs/wazero/api"
	"go.uber.org/zap"
)

func hostGetHeader(ctx context.Context, mod api.Module, keyPtr, keyLen, bufPtr, bufMaxLen uint32) uint32 {
	logger := reqctx.LoggerFromContext(ctx)

	req, ok := reqctx.RequestFromContext(ctx)
	if !ok {
		return 0xFFFFFFFF
	}

	keyBytes, ok := mod.Memory().Read(keyPtr, keyLen)
	if !ok {
		logger.Error("WASM Access Violation: Failed to read header key",
			zap.Uint32("ptr", keyPtr),
			zap.Uint32("len", keyLen),
		)
		return 0xFFFFFFFF
	}
	key := string(keyBytes)

	headerValue := req.Header.Get(key)
	if headerValue == "" {
		return 0xFFFFFFFF
	}

	valBytes := []byte(headerValue)
	writeLen := uint32(len(valBytes))

	if writeLen > bufMaxLen {
		writeLen = bufMaxLen // Truncate if the header value exceeds the buffer size
	}

	if !mod.Memory().Write(bufPtr, valBytes[:writeLen]) {
		logger.Error("WASM Access Violation: Failed ot write header value",
			zap.String("key", key),
		)
	}
	return writeLen
}

func hostSetHeader(ctx context.Context, mod api.Module, keyPtr, keyLen, valPtr, valLen uint32) {
	logger := reqctx.LoggerFromContext(ctx)

	req, ok := reqctx.RequestFromContext(ctx)
	if !ok {
		return
	}

	keyBytes, ok := mod.Memory().Read(keyPtr, keyLen)
	if !ok {
		logger.Error("WASM Access Violation: Failed to read header key",
			zap.Uint32("ptr", keyPtr),
			zap.Uint32("len", keyLen),
		)
		return
	}

	valBytes, ok := mod.Memory().Read(valPtr, valLen)
	if !ok {
		logger.Error("WASM Access Violation: Failed to read header value",
			zap.Uint32("ptr", valPtr),
			zap.Uint32("len", valLen),
		)
		return
	}

	headerName := string(keyBytes)
	headerValue := string(valBytes)

	req.Header.Set(headerName, headerValue)
}

func hostGetMethod(ctx context.Context, mod api.Module, bufPtr, bufMaxLen uint32) uint32 {
	logger := reqctx.LoggerFromContext(ctx)

	req, ok := reqctx.RequestFromContext(ctx)
	if !ok {
		return 0xFFFFFFFF
	}

	methodBytes := []byte(req.Method)
	methodLen := uint32(len(methodBytes))
	if methodLen > bufMaxLen {
		methodLen = bufMaxLen // Truncate if the method exceeds the buffer size
	}

	if !mod.Memory().Write(bufPtr, methodBytes[:methodLen]) {
		logger.Error("WASM Access Violation: Failed to write request method",
			zap.Uint32("buffer_ptr", bufPtr),
			zap.Uint32("len", methodLen),
		)
	}
	return methodLen
}

func hostGetPath(ctx context.Context, mod api.Module, bufPtr, bufMaxLen uint32) uint32 {
	logger := reqctx.LoggerFromContext(ctx)

	req, ok := reqctx.RequestFromContext(ctx)
	if !ok {
		return 0xFFFFFFFF
	}

	pathBytes := []byte(req.URL.Path)
	pathLen := uint32(len(pathBytes))

	if pathLen > bufMaxLen {
		pathLen = bufMaxLen // Truncate if the path exceeds the buffer size
	}

	if !mod.Memory().Write(bufPtr, pathBytes[:pathLen]) {
		logger.Error("WASM Access Violation: Failed to write path",
			zap.Uint32("buffer_ptr", bufPtr),
			zap.Uint32("len", pathLen),
		)
	}
	return pathLen
}

func hostGetQueryParam(ctx context.Context, mod api.Module, keyPtr, keyLen, bufPtr, bufMaxLen uint32) uint32 {
	logger := reqctx.LoggerFromContext(ctx)

	req, ok := reqctx.RequestFromContext(ctx)
	if !ok {
		return 0xFFFFFFFF
	}

	keyBytes, ok := mod.Memory().Read(keyPtr, keyLen)
	if !ok {
		logger.Error("WASM Access Violation: Failed to read key",
			zap.Uint32("ptr", keyPtr),
			zap.Uint32("len", keyLen),
		)
		return 0xFFFFFFFF
	}
	key := string(keyBytes)

	value := req.URL.Query().Get(key)
	if value == "" {
		return 0xFFFFFFFF
	}
	valueBytes := []byte(value)
	valueLen := uint32(len(valueBytes))

	if valueLen > bufMaxLen {
		valueLen = bufMaxLen
	}

	if !mod.Memory().Write(bufPtr, valueBytes[:valueLen]) {
		logger.Error("WASM Access Violation: Failed to write value",
			zap.Uint32("buffer_ptr", bufPtr),
			zap.Uint32("len", valueLen),
		)
		return 0xFFFFFFFF
	}
	return valueLen
}

func hostGetRawQuery(ctx context.Context, mod api.Module, bufPtr, bufMaxLen uint32) uint32 {
	logger := reqctx.LoggerFromContext(ctx)

	req, ok := reqctx.RequestFromContext(ctx)
	if !ok {
		return 0xFFFFFFFF
	}

	queryBytes := []byte(req.URL.RawQuery)
	queryLen := uint32(len(queryBytes))

	if queryLen > bufMaxLen {
		queryLen = bufMaxLen
	}

	if !mod.Memory().Write(bufPtr, queryBytes[:queryLen]) {
		logger.Error("WASM Access Violation: Failed to write raw query",
			zap.Uint32("buffer_ptr", bufPtr),
			zap.Uint32("len", queryLen),
		)
		return 0xFFFFFFFF
	}
	return queryLen
}

func hostSendResponse(ctx context.Context, mod api.Module, statusCode uint32, bodyPtr, bodyLen uint32) {
	logger := reqctx.LoggerFromContext(ctx)

	state := reqctx.RequestStateFromContext(ctx)

	bodyBytes, ok := mod.Memory().Read(bodyPtr, bodyLen)
	if !ok {
		logger.Error("WASM Access Violation: Failed to read body",
			zap.Uint32("ptr", bodyPtr),
			zap.Uint32("len", bodyLen),
		)
		return
	}

	if statusCode >= 200 && statusCode < 300 {
		logger.Debug("WASM module intercepted request and returned success (2XX). Upstream will not be called.", zap.Uint32("status_code", statusCode))
	}

	state.Interrupted = true
	state.StatusCode = int(statusCode)

	state.Body = make([]byte, len(bodyBytes))
	copy(state.Body, bodyBytes)
}

func hostLog(ctx context.Context, mod api.Module, level uint32, msgPtr, msgLen uint32) {
	logger := reqctx.LoggerFromContext(ctx)

	msgBytes, ok := mod.Memory().Read(msgPtr, msgLen)
	if !ok {
		logger.Error("WASM tried to log but pointer was out of bounds")
		return
	}

	message := string(msgBytes)

	switch level {
	case 0:
		logger.Debug(message)
	case 1:
		logger.Info(message)
	case 2:
		logger.Warn(message)
	case 3:
		logger.Error(message)
	default:
		logger.Info(message) // Default to info if level is unrecognized
	}
}

func hostGetJSONConfig(ctx context.Context, mod api.Module, bufPtr, bufMaxLen uint32) uint32 {
	logger := reqctx.LoggerFromContext(ctx)

	configPtr, ok := reqctx.JSONConfigFromContext(ctx)
	if !ok || configPtr == nil {
		return 0xFFFFFFFF
	}

	configBytes := []byte(*configPtr)
	configLen := uint32(len(configBytes))

	if configLen > bufMaxLen {
		configLen = bufMaxLen // Truncate if the config exceeds the buffer size
	}

	if !mod.Memory().Write(bufPtr, configBytes[:configLen]) {
		logger.Error("WASM Access Violation: Failed to write JSON config",
			zap.Uint32("buffer_ptr", bufPtr),
			zap.Uint32("len", configLen),
		)
		return 0xFFFFFFFF
	}
	return configLen
}

func hostAuthIsAuthenticated(ctx context.Context, mod api.Module) uint32 {
	state := reqctx.RequestStateFromContextSafe(ctx)
	if state == nil || state.AuthContext == nil || !state.AuthContext.IsAuthenticated {
		return 0
	}
	return 1
}

func hostAuthSubject(ctx context.Context, mod api.Module, bufPtr, bufMaxLen uint32) uint32 {
	state := reqctx.RequestStateFromContextSafe(ctx)
	if state == nil || state.AuthContext == nil || state.AuthContext.Subject == "" {
		return 0xFFFFFFFF
	}
	return writeStringToMemory(ctx, mod, state.AuthContext.Subject, bufPtr, bufMaxLen, "auth subject")
}

func hostAuthClaim(ctx context.Context, mod api.Module, keyPtr, keyLen, bufPtr, bufMaxLen uint32) uint32 {
	state := reqctx.RequestStateFromContextSafe(ctx)
	logger := reqctx.LoggerFromContext(ctx)
	if state == nil || state.AuthContext == nil || state.AuthContext.ValidatedToken == nil {
		return 0xFFFFFFFF
	}

	keyBytes, ok := mod.Memory().Read(keyPtr, keyLen)
	if !ok {
		logger.Error("WASM Access Violation: Failed to read auth claim key", zap.Uint32("ptr", keyPtr), zap.Uint32("len", keyLen))
		return 0xFFFFFFFF
	}
	claim, exists := state.AuthContext.ValidatedToken.Claims[string(keyBytes)]
	if !exists {
		return 0xFFFFFFFF
	}

	value := ""
	if stringValue, ok := claim.(string); ok {
		value = stringValue
	} else {
		raw, err := json.Marshal(claim)
		if err != nil {
			value = fmt.Sprint(claim)
		} else {
			value = string(raw)
		}
	}
	return writeStringToMemory(ctx, mod, value, bufPtr, bufMaxLen, "auth claim")
}

func writeStringToMemory(ctx context.Context, mod api.Module, value string, bufPtr, bufMaxLen uint32, field string) uint32 {
	logger := reqctx.LoggerFromContext(ctx)
	raw := []byte(value)
	writeLen := uint32(len(raw))
	if writeLen > bufMaxLen {
		writeLen = bufMaxLen
	}
	if !mod.Memory().Write(bufPtr, raw[:writeLen]) {
		logger.Error("WASM Access Violation: Failed to write value", zap.String("field", field), zap.Uint32("buffer_ptr", bufPtr), zap.Uint32("len", writeLen))
		return 0xFFFFFFFF
	}
	return writeLen
}
