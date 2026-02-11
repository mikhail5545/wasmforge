package host

import (
	"context"
	"log"
	"os"

	"github.com/tetratelabs/wazero"
)

func NewWasmRuntime(ctx context.Context) (wazero.Runtime, func() error, error) {
	cacheDir, err := os.MkdirTemp("", "wasm")
	if err != nil {
		log.Printf("Error creating cache directory: %v", err)
	}
	cache, err := wazero.NewCompilationCacheWithDir(cacheDir)
	if err != nil {
		log.Printf("Error creating cache directory: %v", err)
	}

	cfg := wazero.NewRuntimeConfig().WithCompilationCache(cache)
	r := wazero.NewRuntimeWithConfig(ctx, cfg)

	_, err = r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(hostGetHeader).Export("host_get_header").
		NewFunctionBuilder().WithFunc(hostGetPath).Export("host_get_path").
		NewFunctionBuilder().WithFunc(hostGetMethod).Export("host_get_method").
		NewFunctionBuilder().WithFunc(hostSetHeader).Export("host_set_header").
		NewFunctionBuilder().WithFunc(hostSendResponse).Export("host_send_response").
		NewFunctionBuilder().WithFunc(hostLog).Export("host_log").
		NewFunctionBuilder().WithFunc(hostGetQueryParam).Export("host_get_query_param").
		NewFunctionBuilder().WithFunc(hostGetRawQuery).Export("host_get_raw_query").
		Instantiate(ctx)

	if err != nil {
		r.Close(ctx)
		return nil, nil, err
	}
	return r, func() error {
		return cache.Close(context.Background())
	}, nil
}
