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

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func fallbackLogDir(fileName string) (*os.File, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	fallback := filepath.Join(home, ".local", "share", "wasmforge")
	if err := os.MkdirAll(fallback, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create fallback log directory '%s': %w", fallback, err)
	}
	fullPath := filepath.Join(fallback, fileName)
	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file in fallback directory '%s': %w", fallback, err)
	}
	return file, nil
}

func openLogFile(logCfg LogConfig) (*os.File, error) {
	var fileName string
	if logCfg.UseTimestamp {
		now := time.Now()
		fileName = fmt.Sprintf("%02d-%s-%d-%02d-%02d-app.log", now.Day(), now.Month().String(), now.Year(), now.Hour(), now.Minute())
	} else {
		fileName = "app.log"
	}

	// Ensure directory exists
	if err := os.MkdirAll(logCfg.Directory, 0o755); err != nil {
		if os.IsPermission(err) {
			// Try fallback directory if permission denied
			file, err := fallbackLogDir(fileName)
			if err != nil {
				return nil, err
			}
			return file, nil
		}
		return nil, fmt.Errorf("failed to create log directory '%s': %w", logCfg.Directory, err)
	}

	fullPath := filepath.Join(logCfg.Directory, fileName)

	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file '%q': %w", fullPath, err)
	}
	return file, nil
}

// newLogger creates a new zap.Logger based on the provided LogConfig.
// Make sure to call the returned cleanup function to close file handles to prevent potential recourse leak.
func newLogger(logCfg LogConfig) (*zap.Logger, func(), error) {
	f, err := openLogFile(logCfg)
	if err != nil {
		return nil, nil, err
	}

	// writers
	consoleWS := zapcore.Lock(os.Stdout)
	fileWS := zapcore.AddSync(f)

	// encoders
	consoleEncCfg := zap.NewDevelopmentEncoderConfig()
	consoleEncCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEnc := zapcore.NewConsoleEncoder(consoleEncCfg)

	fileEnc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEnc, consoleWS, normalizeLogLevel(logCfg.ConsoleLevel)),
		zapcore.NewCore(fileEnc, fileWS, normalizeLogLevel(logCfg.ConsoleLevel)),
	)

	logger := zap.New(core, zap.AddCaller())
	cleanup := func() {
		_ = logger.Sync()
		_ = f.Close()
	}
	return logger, cleanup, nil
}

// normalizeLogLevel converts a string log level from configuration to a zapcore.Level. It defaults to InfoLevel if the input is unrecognized.
// Case-insensitive.
func normalizeLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel // default to info if unrecognized
	}
}
