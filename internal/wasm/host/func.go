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

package host

import (
	"context"

	"github.com/mikhail5545/wasm-gateway/internal/reqctx"
	"github.com/tetratelabs/wazero/api"
	"go.uber.org/zap"
)

func hostGetHeader(ctx context.Context, mod api.Module, keyPtr, keyLen, bufPtr, bufMaxLen uint32) uint32 {
	logger := reqctx.LoggerFromContext(ctx)

	req, ok := reqctx.RequestFromContext(ctx)
	if !ok {
		return 0
	}

	keyBytes, ok := mod.Memory().Read(keyPtr, keyLen)
	if !ok {
		logger.Error("WASM Access Violation: Failed to read header key",
			zap.Uint32("ptr", keyPtr),
			zap.Uint32("len", keyLen),
		)
		return 0
	}
	key := string(keyBytes)

	headerValue := req.Header.Get(key)
	if headerValue == "" {
		return 0
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
		return 0
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
		return 0
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
		return 0
	}

	keyBytes, ok := mod.Memory().Read(keyPtr, keyLen)
	if !ok {
		logger.Error("WASM Access Violation: Failed to read key",
			zap.Uint32("ptr", keyPtr),
			zap.Uint32("len", keyLen),
		)
		return 0
	}
	key := string(keyBytes)

	valueBytes := []byte(req.URL.Query().Get(key))
	valueLen := uint32(len(valueBytes))

	if valueLen > bufMaxLen {
		valueLen = bufMaxLen
	}

	if !mod.Memory().Write(bufPtr, valueBytes[:valueLen]) {
		logger.Error("WASM Access Violation: Failed to write value",
			zap.Uint32("buffer_ptr", bufPtr),
			zap.Uint32("len", valueLen),
		)
		return 0
	}
	return valueLen
}

func hostGetRawQuery(ctx context.Context, mod api.Module, bufPtr, bufMaxLen uint32) uint32 {
	logger := reqctx.LoggerFromContext(ctx)

	req, ok := reqctx.RequestFromContext(ctx)
	if !ok {
		return 0
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
		return 0
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
