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
