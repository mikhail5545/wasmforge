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

package uploads

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"

	inerrors "github.com/mikhail5545/wasmforge/internal/errors"

	"go.uber.org/zap"
)

type Manager struct {
	pluginUploadDir string
	logger          *zap.Logger
}

func New(pluginUploadDir string, logger *zap.Logger) *Manager {
	return &Manager{
		pluginUploadDir: pluginUploadDir,
		logger:          logger.With(zap.String("component", "uploads_manager")),
	}
}

func (m *Manager) PluginUploadDir() string {
	return m.pluginUploadDir
}

func (m *Manager) EnsureDirectory() error {
	_, err := os.ReadDir(m.pluginUploadDir)
	if err != nil {
		if os.IsNotExist(err) {
			m.logger.Info("uploads directory does not exist, creating it", zap.String("path", m.pluginUploadDir))
			if err := os.MkdirAll(m.pluginUploadDir, 0755); err != nil {
				m.logger.Error("failed to create uploads directory", zap.String("path", m.pluginUploadDir), zap.Error(err))
				return err
			}
			m.logger.Info("uploads directory created successfully", zap.String("path", m.pluginUploadDir))
			return nil
		}
	}
	return nil
}

func (m *Manager) FromBase64(encodedData, filename string) (string, string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		m.logger.Error("failed to decode Base64 data", zap.String("filename", filename), zap.Error(err))
		return "", "", fmt.Errorf("failed to decode Base64 data: %w", err)
	}

	if len(decoded) > 10*1024*1024 { // 10 MB limit
		m.logger.Warn("decoded data exceeds size limit", zap.String("filename", filename), zap.Int("size", len(decoded)))
		return "", "", inerrors.NewSizeLimitExceededError("file size exceedes the limit of 10 MB")
	}

	filePath := path.Join(m.pluginUploadDir, filename)
	if err := os.WriteFile(filePath, decoded, 0644); err != nil {
		m.logger.Error("failed to write decoded data to file", zap.String("filename", filePath), zap.Error(err))
		return "", "", fmt.Errorf("failed to write decoded data to file: %w", err)
	}

	hash := hashFromBytes(decoded)

	m.logger.Debug("file created successfully from Base64 data", zap.String("filename", filePath), zap.Int("size", len(decoded)), zap.String("hash", hash))
	return filePath, hash, nil
}

func (m *Manager) FromMultipartFile(file *multipart.FileHeader, filename string) (string, error) {
	src, err := file.Open()
	if err != nil {
		m.logger.Error("failed to open source file", zap.String("filename", file.Filename), zap.Error(err))
		return "", err
	}
	defer func() {
		if err := src.Close(); err != nil {
			m.logger.Warn("failed to close source file", zap.String("filename", file.Filename), zap.Error(err))
		}
	}()

	dstPath := path.Join(m.pluginUploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		m.logger.Error("failed to create destination file", zap.String("filename", dstPath), zap.Error(err))
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if err := dst.Close(); err != nil {
			m.logger.Warn("failed to close destination file", zap.String("filename", dstPath), zap.Error(err))
		}
	}()

	m.logger.Debug("starting to copy file", zap.String("source", file.Filename), zap.String("destination", dstPath))
	if _, err := io.Copy(dst, src); err != nil {
		m.logger.Error("failed to copy file", zap.String("source", file.Filename), zap.String("destination", dstPath), zap.Error(err))
		return "", fmt.Errorf("failed to copy file: %w", err)
	}
	fileHash, err := hashFromOpen(src)
	if err != nil {
		m.logger.Error("failed to compute file hash", zap.String("filename", file.Filename), zap.Error(err))
		return "", fmt.Errorf("failed to compute file hash: %w", err)
	}
	m.logger.Debug("file saved successfully", zap.String("filename", dstPath), zap.String("hash", fileHash))
	return fileHash, nil
}

func (m *Manager) Delete(filename string) error {
	filePath := path.Join(m.pluginUploadDir, filename)
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			m.logger.Warn("file does not exist, nothing to delete", zap.String("filename", filePath))
			return nil
		}
		m.logger.Error("failed to delete file", zap.String("filename", filePath), zap.Error(err))
		return fmt.Errorf("failed to delete file: %w", err)
	}
	m.logger.Debug("file deleted successfully", zap.String("filename", filePath))
	return nil
}

func (m *Manager) Read(filename string) ([]byte, error) {
	filePath := path.Join(m.pluginUploadDir, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}
