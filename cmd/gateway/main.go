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

package main

import (
	"context"
	"log"
	"os"

	"github.com/mikhail5545/wasm-gateway/internal/proxy"
	"github.com/mikhail5545/wasm-gateway/internal/wasm/host"
	"github.com/mikhail5545/wasm-gateway/internal/wasm/middleware"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}

	runtime, cleanup, err := host.NewWasmRuntime(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			logger.Error("failed to clean up WASM runtime cache", zap.Error(err))
		}
		if err := runtime.Close(ctx); err != nil {
			logger.Error("failed to close WASM runtime", zap.Error(err))
		}
	}()

	wasmBytes, err := os.ReadFile("./plugins/auth.wasm")
	if err != nil {
		log.Fatal("Could not read WASM file:", err)
	}

	wasmMiddleware, err := middleware.New(ctx, runtime, wasmBytes, logger)
	if err != nil {
		log.Fatal("Failed to create WASM middleware:", err)
	}

	proxyManager := proxy.New()
	proxyManager.AddRoute("/api/", "http://localhost:8081", wasmMiddleware)

}
