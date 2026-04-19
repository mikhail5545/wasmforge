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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/mikhail5545/wasmforge/internal/crypto"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../mocks/uploads/manager.go -package=uploads . Manager

type Manager interface {
	PluginUploadDir() string
	CertUploadDir() string
	EnsureDirectory(uploadType UploadType) error
	FromBase64(encodedData, filename string, uploadType UploadType) (string, error)
	FromBytes(data []byte, filename string, uploadType UploadType) (string, error)
	FromMultipartFile(file *multipart.FileHeader, filename string, uploadType UploadType) (string, error)
	Delete(filename string, uploadType UploadType) error
	Read(filename string, uploadType UploadType) ([]byte, error)
}

type manager struct {
	pluginUploadDir string
	certUploadDir   string
	logger          *zap.Logger
}

type UploadType uint

const (
	PluginUpload UploadType = iota
	CertUpload
)

const (
	maxUploadSizeBytes = 100 * 1024 * 1024
	maxUploadSizeLabel = "100 MB"
)

func New(pluginUploadDir string, certUploadDir string, logger *zap.Logger) Manager {
	return &manager{
		pluginUploadDir: pluginUploadDir,
		certUploadDir:   certUploadDir,
		logger:          logger.With(zap.String("component", "uploads_manager")),
	}
}

func (m *manager) PluginUploadDir() string {
	return m.pluginUploadDir
}

func (m *manager) CertUploadDir() string {
	return m.certUploadDir
}

func (m *manager) EnsureDirectory(uploadType UploadType) error {
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

func (m *manager) FromBase64(encodedData, filename string, uploadType UploadType) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		m.logger.Error("failed to decode Base64 data", zap.String("filename", filename), zap.Error(err))
		return "", fmt.Errorf("failed to decode Base64 data: %w", err)
	}

	if len(decoded) > maxUploadSizeBytes {
		m.logger.Warn("decoded data exceeds size limit", zap.String("filename", filename), zap.Int("size", len(decoded)))
		return "", inerrors.NewSizeLimitExceededError("file size exceeds the limit of " + maxUploadSizeLabel)
	}

	fullPath, err := m.buildFullPath(uploadType, filename)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(fullPath, decoded, 0644); err != nil {
		m.logger.Error("failed to write decoded data to file", zap.String("filename", fullPath), zap.Error(err))
		return "", fmt.Errorf("failed to write decoded data to file: %w", err)
	}

	hash := crypto.HashFromBytes(decoded)

	m.logger.Debug("file created successfully from Base64 data", zap.String("filename", fullPath), zap.Int("size", len(decoded)), zap.String("hash", hash))
	return hash, nil
}

func (m *manager) FromBytes(data []byte, filename string, uploadType UploadType) (string, error) {
	if len(data) > maxUploadSizeBytes {
		m.logger.Warn("data exceeds size limit", zap.String("filename", filename), zap.Int("size", len(data)))
		return "", inerrors.NewSizeLimitExceededError("file size exceeds the limit of " + maxUploadSizeLabel)
	}

	fullPath, err := m.buildFullPath(uploadType, filename)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		m.logger.Error("failed to write data to file", zap.String("filename", fullPath), zap.Error(err))
		return "", fmt.Errorf("failed to write data to file: %w", err)
	}

	hash := crypto.HashFromBytes(data)

	m.logger.Debug("file created successfully from bytes data", zap.String("filename", fullPath), zap.Int("size", len(data)), zap.String("hash", hash))
	return hash, nil
}

func (m *manager) FromMultipartFile(file *multipart.FileHeader, filename string, uploadType UploadType) (string, error) {
	if file.Size > maxUploadSizeBytes {
		m.logger.Warn("multipart data exceeds size limit", zap.String("filename", filename), zap.Int64("size", file.Size))
		return "", inerrors.NewSizeLimitExceededError("file size exceeds the limit of " + maxUploadSizeLabel)
	}

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
	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(dst, hasher), io.LimitReader(src, maxUploadSizeBytes+1))
	if err != nil {
		m.logger.Error("failed to copy file", zap.String("source", file.Filename), zap.String("destination", dstPath), zap.Error(err))
		return "", fmt.Errorf("failed to copy file: %w", err)
	}
	if written > maxUploadSizeBytes {
		m.logger.Warn("multipart data exceeded size limit while streaming", zap.String("filename", filename), zap.Int64("size", written))
		if removeErr := os.Remove(dstPath); removeErr != nil {
			m.logger.Warn("failed to remove oversized destination file", zap.String("filename", dstPath), zap.Error(removeErr))
		}
		return "", inerrors.NewSizeLimitExceededError("file size exceeds the limit of " + maxUploadSizeLabel)
	}

	fileHash := hex.EncodeToString(hasher.Sum(nil))
	m.logger.Debug("file saved successfully", zap.String("filename", dstPath), zap.String("hash", fileHash))
	return fileHash, nil
}

func (m *manager) Delete(filename string, uploadType UploadType) error {
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

func (m *manager) Read(filename string, uploadType UploadType) ([]byte, error) {
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
