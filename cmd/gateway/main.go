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

	"github.com/mikhail5545/wasmforge/internal/app"
	"github.com/spf13/pflag"
)

func main() {
	ctx := context.Background()

	cfg := parseArgs()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}
	defer func(application *app.App, ctx context.Context) {
		err := application.Cleanup(ctx)
		if err != nil {
			log.Fatalf("failed to cleanup app: %v", err)
		}
	}(application, ctx)

	if err := application.Init(ctx); err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	if err := application.Start(ctx); err != nil {
		log.Fatalf("app stopped with error: %v", err)
	}
}

func parseArgs() *app.Config {
	cfg := &app.Config{
		DatabaseConfig: app.DatabaseConfig{
			DSN: "./wasmforge.db",
		},
	}

	pflag.Int64VarP(&cfg.AdminServerConfig.Port, "admin-port", "a", 8080, "Port for the admin server")
	pflag.Int64VarP(&cfg.ProxyServerConfig.Port, "proxy-port", "p", 9090, "Port for the proxy server")
	pflag.StringVarP(&cfg.UploadsConfig.Directory, "uploads-dir", "u", "./uploads", "Directory for uploaded WASM modules")
	pflag.StringVarP(&cfg.LogConfig.Directory, "logs-dir", "l", "./logs", "Directory for log files")
	pflag.BoolVarP(&cfg.LogConfig.UseTimestamp, "use-timestamps", "t", true, "Use timestamps in logs filenames")
	pflag.Parse()

	return cfg
}
