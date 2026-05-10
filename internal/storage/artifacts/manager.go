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

package artifacts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/google/uuid"
	artifactrepo "github.com/mikhail5545/wasmforge/internal/database/storage/artifact"
	artifactmodel "github.com/mikhail5545/wasmforge/internal/models/storage/artifact"
	runtime "github.com/mikhail5545/wasmforge/internal/runtime/core"
	"github.com/mikhail5545/wasmforge/internal/storage/core"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	UploadParams struct {
		UploadName string
		Version    string
		Role       artifactmodel.Role
		Entrypoint string

		AppID     *uuid.UUID
		ProjectID uuid.UUID
		R         io.Reader
		SizeBytes int64
	}

	UploadFromPathParams struct {
		UploadName string
		Version    string
		Role       artifactmodel.Role
		Entrypoint string

		AppID     *uuid.UUID
		ProjectID uuid.UUID
		Path      string
		SizeBytes int64
	}

	UploadFromMultipartParams struct {
		UploadName string
		Version    string
		Role       artifactmodel.Role
		Entrypoint string

		AppID     *uuid.UUID
		ProjectID uuid.UUID
		File      *multipart.FileHeader
		SizeBytes int64
	}

	UploadFromBytesParams struct {
		UploadName string
		Version    string
		Role       artifactmodel.Role
		Entrypoint string

		AppID     *uuid.UUID
		ProjectID uuid.UUID
		Data      []byte
		SizeBytes int64
	}

	ArtifactInfo struct {
		ID                uuid.UUID
		Name              string
		ProjectID         uuid.UUID
		AppID             *uuid.UUID
		Version           string
		Role              artifactmodel.Role
		Status            artifactmodel.Status
		SizeBytes         int64
		ChecksumSHA256Hex string
		Metadata          Metadata
	}

	Config struct {
		DataRoot           string
		MaxUploadSizeBytes int64
	}

	ManagerParams struct {
		ArtifactRepo artifactrepo.Repository
		ObjectStore  core.ObjectStore
	}
)

type Manager interface {
	UploadFromMultipart(ctx context.Context, params UploadFromMultipartParams) (ArtifactInfo, error)
	UploadFromPath(ctx context.Context, params UploadFromPathParams) (ArtifactInfo, error)
	UploadFromBytes(ctx context.Context, params UploadFromBytesParams) (ArtifactInfo, error)
	Upload(ctx context.Context, params UploadParams) (ArtifactInfo, error)
}

type manager struct {
	cfg          Config
	validator    *Validator
	refBuilder   *RefBuilder
	artifactRepo artifactrepo.Repository
	objectStore  core.ObjectStore
	logger       *zap.Logger
}

func NewManager(cfg Config, params ManagerParams, runtime runtime.Runtime, logger *zap.Logger) Manager {
	return &manager{
		cfg:          cfg,
		validator:    NewValidator(runtime, logger),
		refBuilder:   NewRefBuilder(cfg.DataRoot),
		objectStore:  params.ObjectStore,
		artifactRepo: params.ArtifactRepo,
		logger:       logger.With(zap.String("domain", "storage"), zap.String("component", "artifact_manager")),
	}
}

func (m *manager) UploadFromMultipart(ctx context.Context, params UploadFromMultipartParams) (ArtifactInfo, error) {
	m.logger.Debug("uploading artifact from multipart file header", zap.String("name", params.UploadName))

	r, err := params.File.Open()
	if err != nil {
		m.logger.Error("failed to open multipart file", zap.Error(err))
		return ArtifactInfo{}, fmt.Errorf("failed to open multipart file: %w", err)
	}
	defer func() {
		if err := r.Close(); err != nil {
			m.logger.Error("failed to close multipart file", zap.Error(err))
		}
	}()

	return m.Upload(ctx, UploadParams{
		UploadName: params.UploadName,
		Version:    params.Version,
		Role:       params.Role,
		Entrypoint: params.Entrypoint,
		AppID:      params.AppID,
		ProjectID:  params.ProjectID,
		R:          r,
		SizeBytes:  params.SizeBytes,
	})
}

func (m *manager) UploadFromPath(ctx context.Context, params UploadFromPathParams) (ArtifactInfo, error) {
	m.logger.Debug("uploading artifact from path header", zap.String("name", params.UploadName))

	file, err := os.Open(params.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			m.logger.Error("provided source file does not exist", zap.String("name", params.UploadName), zap.String("path", params.Path), zap.Error(err))
			return ArtifactInfo{}, fmt.Errorf("provided source file does not exist: %w", err)
		}
		m.logger.Error("failed to open source file", zap.Error(err))
		return ArtifactInfo{}, fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			m.logger.Error("failed to close source file", zap.Error(err))
		}
	}()

	return m.Upload(ctx, UploadParams{
		UploadName: params.UploadName,
		Version:    params.Version,
		Role:       params.Role,
		Entrypoint: params.Entrypoint,
		AppID:      params.AppID,
		ProjectID:  params.ProjectID,
		R:          file,
	})
}

func (m *manager) UploadFromBytes(ctx context.Context, params UploadFromBytesParams) (ArtifactInfo, error) {
	m.logger.Debug("uploading artifact from bytes header", zap.String("name", params.UploadName))

	return m.Upload(ctx, UploadParams{
		UploadName: params.UploadName,
		Version:    params.Version,
		Role:       params.Role,
		Entrypoint: params.Entrypoint,
		AppID:      params.AppID,
		ProjectID:  params.ProjectID,
		R:          bytes.NewReader(params.Data),
	})
}

func (m *manager) Upload(ctx context.Context, params UploadParams) (ArtifactInfo, error) {
	m.logger.Debug("uploading artifact", zap.String("name", params.UploadName))

	if params.SizeBytes > m.cfg.MaxUploadSizeBytes {
		return ArtifactInfo{}, core.NewSizeLimitExceededError(fmt.Sprintf("file size %d exceeds the maximum allowed size of %d bytes", params.SizeBytes, m.cfg.MaxUploadSizeBytes))
	}

	tempRef, err := m.refBuilder.BuildTemp()
	if err != nil {
		return ArtifactInfo{}, err
	}

	dest, err := os.Create(tempRef.Key)
	if err != nil {
		m.logger.Error("failed to create temp file", zap.Error(err))
		return ArtifactInfo{}, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		if err := dest.Close(); err != nil {
			m.logger.Error("failed to close temp file", zap.Error(err))
		}
	}()

	_, err = io.Copy(dest, params.R)
	if err != nil {
		if err := dest.Close(); err != nil {
			m.logger.Error("failed to close temp file", zap.Error(err))
		}
		return ArtifactInfo{}, fmt.Errorf("failed to write to temp file: %w", err)
	}
	m.logger.Debug("saved file to temporary storage", zap.String("key", tempRef.Key))
	if err := dest.Close(); err != nil {
		m.logger.Error("failed to close temp file", zap.Error(err))
		return ArtifactInfo{}, fmt.Errorf("failed to close temp file: %w", err)
	}

	// Reopen for reading
	readDest, err := os.Open(tempRef.Key)
	if err != nil {
		m.logger.Error("failed to open temp file", zap.Error(err))
		return ArtifactInfo{}, fmt.Errorf("failed to open temp file: %w", err)
	}
	defer func() {
		if err := readDest.Close(); err != nil {
			m.logger.Error("failed to close temp file", zap.Error(err))
		}
	}()

	validationInput := ValidationInput{
		Name:     params.UploadName,
		R:        readDest,
		SizeHint: params.SizeBytes,
		Ref:      tempRef,
	}
	validationRes, err := m.validator.Validate(ctx, validationInput)
	if err != nil {
		m.logger.Error("validation returned an error", zap.Error(err))
		return ArtifactInfo{}, err
	}

	m.logger.Debug("validation successful, proceeding to moving file to final location")

	return m.saveArtifact(ctx, validationRes, params, tempRef)
}

func (m *manager) saveArtifact(ctx context.Context, validationRes ValidationResult, params UploadParams, tempRef core.ObjectRef) (ArtifactInfo, error) {
	artifactID, err := uuid.NewV7()
	if err != nil {
		m.logger.Error("failed to generate artifact ID", zap.Error(err))
		return ArtifactInfo{}, fmt.Errorf("failed to generate artifact ID: %w", err)
	}

	metadata, err := json.Marshal(validationRes.Metadata)
	if err != nil {
		m.logger.Error("failed to marshal artifact metadata", zap.Error(err))
		return ArtifactInfo{}, fmt.Errorf("failed to marshal artifact metadata: %w", err)
	}

	savedRef := m.refBuilder.Build(params.ProjectID, artifactID, params.AppID)

	// Reopen for reading
	file, err := os.Open(tempRef.Key)
	if err != nil {
		m.logger.Error("failed to open temp file", zap.Error(err))
		return ArtifactInfo{}, fmt.Errorf("failed to open temp file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			m.logger.Error("failed to close temp file", zap.Error(err))
		}
	}()

	info, err := m.objectStore.Put(ctx, savedRef, file)
	if err != nil {
		m.logger.Error("failed to save artifact", zap.Error(err))
		return ArtifactInfo{}, fmt.Errorf("failed to save artifact: %w", err)
	}

	artifact := &artifactmodel.Artifact{
		ID:                artifactID,
		AppID:             params.AppID,
		ProjectID:         params.ProjectID,
		ChecksumSHA256Hex: validationRes.ChecksumSHA256Hex,
		ObjectRefBucket:   savedRef.Bucket.String(),
		ObjectRefKey:      savedRef.Key,
		SizeBytes:         info.SizeBytes,
		Role:              params.Role,
		Version:           params.Version,
		Status:            artifactmodel.StatusValidated,
		Metadata:          string(metadata),
	}

	m.logger.Debug("saving artifact info to the database", zap.String("name", params.UploadName))
	err = m.artifactRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := m.artifactRepo.WithTx(tx)

		if err := txRepo.Create(ctx, artifact); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				m.logger.Debug("artifact already exists", zap.String("key", savedRef.Key))
				return core.NewDuplicateInputError("artifact with the same version, project and name already exists")
			}
			m.logger.Error("failed to create artifact", zap.Error(err))
			return fmt.Errorf("failed to create artifact: %w", err)
		}

		return nil
	})
	if err != nil {
		// Delete file in case of any error to avoid orphaned files
		if savedRef.Key != "" {
			_ = m.objectStore.Delete(ctx, savedRef)
		}
		return ArtifactInfo{}, err
	}
	return ArtifactInfo{
		ID:                artifactID,
		Name:              artifact.Name,
		ProjectID:         artifact.ProjectID,
		AppID:             artifact.AppID,
		Version:           artifact.Version,
		Status:            artifact.Status,
		SizeBytes:         artifact.SizeBytes,
		ChecksumSHA256Hex: artifact.ChecksumSHA256Hex,
		Role:              artifact.Role,
		Metadata:          validationRes.Metadata,
	}, nil
}
