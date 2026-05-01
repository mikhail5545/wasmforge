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
	pflag.StringVarP(&cfg.UploadsConfig.PluginsDirectory, "plugins-uploads-dir", "p", "./uploads", "Directory for uploaded WASM modules")
	pflag.StringVarP(&cfg.UploadsConfig.CertsDirectory, "certs-uploads-dir", "c", "./certs", "Directory for uploaded TLS certificates")
	pflag.StringVarP(&cfg.LogConfig.Directory, "logs-dir", "l", "./logs", "Directory for log files")
	pflag.BoolVarP(&cfg.LogConfig.UseTimestamp, "use-timestamps", "t", true, "Use timestamps in logs filenames")
	pflag.StringVarP(&cfg.LogConfig.FileLevel, "file-log-level", "f", "info", "Case-insensitive log level for file output (debug, info, warn, error)")
	pflag.StringVarP(&cfg.LogConfig.ConsoleLevel, "console-log-level", "s", "debug", "Case-insensitive log level for console output (debug, info, warn, error)")
	pflag.StringVar(&cfg.AuthEncryption.Provider, "auth-encryption-provider", "local", "Auth key encryption provider (local, 1password, or aws-kms)")
	pflag.StringVar(&cfg.AuthEncryption.MasterKeyEnvName, "auth-encryption-master-key-env", "WASMFORGE_AUTH_MASTER_KEY", "Environment variable containing the local auth encryption master key")
	pflag.StringVar(&cfg.AuthEncryption.OnePasswordReference, "auth-encryption-1password-reference", "", "1Password secret reference for the auth encryption master key")
	pflag.StringVar(&cfg.AuthEncryption.OnePasswordTokenEnv, "auth-encryption-1password-token-env", "OP_SERVICE_ACCOUNT_TOKEN", "Environment variable containing the 1Password service account token")
	pflag.StringVar(&cfg.AuthEncryption.OnePasswordIntegration, "auth-encryption-1password-integration", "wasmforge", "Integration name reported to the 1Password SDK")
	pflag.StringVar(&cfg.AuthEncryption.AWSKMSRegion, "auth-encryption-aws-kms-region", "", "AWS region for the auth encryption KMS key")
	pflag.StringVar(&cfg.AuthEncryption.AWSKMSKeyID, "auth-encryption-aws-kms-key-id", "", "AWS KMS key ID or ARN for auth encryption")
	pflag.Parse()

	return cfg
}
