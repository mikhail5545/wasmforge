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

	inerrors "github.com/mikhail5545/wasmforge/internal/errors"

	"go.uber.org/zap"
)

type Manager struct {
	pluginUploadDir string
	certUploadDir   string
	logger          *zap.Logger
}

type UploadType uint

const (
	PluginUpload UploadType = iota
	CertUpload
)

func New(pluginUploadDir string, certUploadDir string, logger *zap.Logger) *Manager {
	return &Manager{
		pluginUploadDir: pluginUploadDir,
		certUploadDir:   certUploadDir,
		logger:          logger.With(zap.String("component", "uploads_manager")),
	}
}

func (m *Manager) PluginUploadDir() string {
	return m.pluginUploadDir
}

func (m *Manager) CertUploadDir() string {
	return m.certUploadDir
}

func (m *Manager) EnsureDirectory(uploadType UploadType) error {
	var dir string
	switch uploadType {
	case PluginUpload:
		dir = m.pluginUploadDir
	case CertUpload:
		dir = m.certUploadDir
	default:
		return fmt.Errorf("invalid upload type: %d", uploadType)
	}

	return m.ensureDirectory(dir)
}

func (m *Manager) FromBase64(encodedData, filename string, uploadType UploadType) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		m.logger.Error("failed to decode Base64 data", zap.String("filename", filename), zap.Error(err))
		return "", fmt.Errorf("failed to decode Base64 data: %w", err)
	}

	if len(decoded) > 10*1024*1024 { // 10 MB limit
		m.logger.Warn("decoded data exceeds size limit", zap.String("filename", filename), zap.Int("size", len(decoded)))
		return "", inerrors.NewSizeLimitExceededError("file size exceeds the limit of 10 MB")
	}

	fullPath, err := m.buildFullPath(uploadType, filename)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(fullPath, decoded, 0644); err != nil {
		m.logger.Error("failed to write decoded data to file", zap.String("filename", fullPath), zap.Error(err))
		return "", fmt.Errorf("failed to write decoded data to file: %w", err)
	}

	hash := hashFromBytes(decoded)

	m.logger.Debug("file created successfully from Base64 data", zap.String("filename", fullPath), zap.Int("size", len(decoded)), zap.String("hash", hash))
	return hash, nil
}

func (m *Manager) FromMultipartFile(file *multipart.FileHeader, filename string, uploadType UploadType) (string, error) {
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

	dstPath, err := m.buildFullPath(uploadType, filename)
	if err != nil {
		return "", err
	}

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

func (m *Manager) Delete(filename string, uploadType UploadType) error {
	fullPath, err := m.buildFullPath(uploadType, filename)
	if err != nil {
		return err
	}
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			m.logger.Warn("file does not exist, nothing to delete", zap.String("filename", fullPath))
			return nil
		}
		m.logger.Error("failed to delete file", zap.String("filename", fullPath), zap.Error(err))
		return fmt.Errorf("failed to delete file: %w", err)
	}
	m.logger.Debug("file deleted successfully", zap.String("filename", fullPath))
	return nil
}

func (m *Manager) Read(filename string, uploadType UploadType) ([]byte, error) {
	fullPath, err := m.buildFullPath(uploadType, filename)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}
