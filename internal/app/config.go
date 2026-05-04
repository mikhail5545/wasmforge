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

package app

import "os"

type (
	LogConfig struct {
		Directory    string
		FileLogs     bool
		UseTimestamp bool
		FileLevel    string
		ConsoleLevel string
	}

	AdminServerConfig struct {
		Port int64
	}

	DatabaseConfig struct {
		DSN string
	}

	UploadsConfig struct {
		PluginsDirectory string
		CertsDirectory   string
	}

	AuthEncryptionConfig struct {
		Provider               string
		MasterKeyEnvName       string
		OnePasswordReference   string
		OnePasswordTokenEnv    string
		OnePasswordIntegration string
		AWSKMSRegion           string
		AWSKMSKeyID            string
	}

	Config struct {
		LogConfig         LogConfig
		AdminServerConfig AdminServerConfig
		DatabaseConfig    DatabaseConfig
		UploadsConfig     UploadsConfig
		AuthEncryption    AuthEncryptionConfig
	}
)

func (c AuthEncryptionConfig) Validate() error {
	switch c.Provider {
	case "", "local":
		envName := c.MasterKeyEnvName
		if envName == "" {
			envName = "WASMFORGE_AUTH_MASTER_KEY"
		}
		if _, ok := lookupNonEmptyEnv(envName); !ok {
			return &ConfigError{Message: "missing auth encryption master key env: " + envName}
		}
		return nil
	case "1password":
		if c.OnePasswordReference == "" {
			return &ConfigError{Message: "1password provider requires a secret reference"}
		}
		tokenEnv := c.OnePasswordTokenEnv
		if tokenEnv == "" {
			tokenEnv = "OP_SERVICE_ACCOUNT_TOKEN"
		}
		if _, ok := lookupNonEmptyEnv(tokenEnv); !ok {
			return &ConfigError{Message: "missing 1password service account token env: " + tokenEnv}
		}
		return nil
	case "aws-kms":
		if c.AWSKMSKeyID == "" {
			return &ConfigError{Message: "aws-kms provider requires a Key ID"}
		}
		return nil
	default:
		return &ConfigError{Message: "unsupported auth encryption provider: " + c.Provider}
	}
}

type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}

func lookupNonEmptyEnv(name string) (string, bool) {
	value, ok := os.LookupEnv(name)
	if !ok || value == "" {
		return "", false
	}
	return value, true
}
