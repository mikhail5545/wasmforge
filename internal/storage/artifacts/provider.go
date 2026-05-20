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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	artifactrepo "github.com/mikhail5545/wasmforge/internal/database/storage/artifact"
	artifactmodel "github.com/mikhail5545/wasmforge/internal/models/storage/artifact"
	"github.com/mikhail5545/wasmforge/internal/storage/core"
	validationutil "github.com/mikhail5545/wasmforge/internal/util/validation"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	// LoadOptions loads an artifact by provided identifier.
	// Must specify one of the following identifiers, otherwise validation will fail:
	//
	//	- ID
	//	- Name, Version and ProjectID
	// 	- Ref
	LoadOptions struct {
		ID        *uuid.UUID
		ProjectID *uuid.UUID
		Name      *string
		Version   *string
		Ref       *core.ObjectRef
	}

	LoadedArtifact struct {
		R        io.ReadCloser
		Artifact ArtifactInfo
		Ref      core.ObjectRef
	}

	ProviderParams struct {
		DataRoot     string
		ArtifactRepo artifactrepo.Repository
		ObjectStore  core.ObjectStore
	}
)

// Provider provides methods to load artifacts from the object store.
// It uses artifact repository to get artifact metadata and object store to get artifact content.
// It is used by the runtime engine to load artifacts for execution.
type Provider interface {
	LoadArtifact(ctx context.Context, options LoadOptions) (LoadedArtifact, error)
	LoadArtifacts(ctx context.Context, appID, projectID *uuid.UUID, ids uuid.UUIDs) ([]LoadedArtifact, error)
}

type provider struct {
	dataRoot     string
	artifactRepo artifactrepo.Repository
	objectStore  core.ObjectStore
	logger       *zap.Logger
}

func NewProvider(params ProviderParams, logger *zap.Logger) Provider {
	return &provider{
		dataRoot:     params.DataRoot,
		artifactRepo: params.ArtifactRepo,
		objectStore:  params.ObjectStore,
		logger:       logger.With(zap.String("domain", "storage"), zap.String("component", "artifact_provider")),
	}
}

func (p *provider) LoadArtifact(ctx context.Context, options LoadOptions) (LoadedArtifact, error) {
	if err := options.Validate(); err != nil {
		return LoadedArtifact{}, fmt.Errorf("invalid options: %w", err)
	}

	artifact, err := p.getArtifact(ctx, options)
	if err != nil {
		return LoadedArtifact{}, err
	}

	var ref core.ObjectRef
	if options.Ref != nil {
		ref = *options.Ref
	} else {
		ref = core.ObjectRef{
			Bucket: core.BucketType(artifact.ObjectRefBucket),
			Key:    artifact.ObjectRefKey,
		}
	}

	r, _, err := p.objectStore.Get(ctx, ref)
	if err != nil {
		p.logger.Error("failed to get artifact from object store", zap.Error(err), zap.Any("object_ref", ref))
		return LoadedArtifact{}, fmt.Errorf("failed to get artifact from object store: %w", err)
	}

	var metadata Metadata
	if err := json.Unmarshal([]byte(artifact.Metadata), &metadata); err != nil {
		return LoadedArtifact{}, fmt.Errorf("failed to unmarshal artifact metadata: %w", err)
	}

	return LoadedArtifact{
		Artifact: ArtifactInfo{
			ID:                artifact.ID,
			ProjectID:         artifact.ProjectID,
			AppID:             artifact.AppID,
			Name:              artifact.Name,
			Version:           artifact.Version,
			Role:              artifact.Role,
			Status:            artifact.Status,
			SizeBytes:         artifact.SizeBytes,
			ChecksumSHA256Hex: artifact.ChecksumSHA256Hex,
			Metadata:          metadata,
		},
		Ref: ref,
		R:   r,
	}, nil
}

func (p *provider) LoadArtifacts(ctx context.Context, appID, projectID *uuid.UUID, ids uuid.UUIDs) ([]LoadedArtifact, error) {
	if appID == nil && projectID == nil && len(ids) == 0 {
		return nil, fmt.Errorf("at least one of appID, projectID or ids must be provided")
	}

	artifacts, err := p.listArtifacts(ctx, appID, projectID, ids)
	if err != nil {
		return nil, err
	}
	if len(artifacts) == 0 {
		return nil, nil
	}

	loaded := make([]LoadedArtifact, 0, len(artifacts))
	for _, artifact := range artifacts {
		ref := core.ObjectRef{
			Bucket: core.BucketType(artifact.ObjectRefBucket),
			Key:    artifact.ObjectRefKey,
		}
		r, _, err := p.objectStore.Get(ctx, ref)
		if err != nil {
			p.logger.Error("failed to get artifact from object store", zap.Error(err), zap.Any("object_ref", ref))
			continue
		}
		var metadata Metadata
		if err := json.Unmarshal([]byte(artifact.Metadata), &metadata); err != nil {
			p.logger.Error("failed to unmarshal artifact metadata", zap.Error(err))
			continue
		}
		loaded = append(loaded, LoadedArtifact{
			Artifact: ArtifactInfo{
				ID:                artifact.ID,
				ProjectID:         artifact.ProjectID,
				AppID:             artifact.AppID,
				Name:              artifact.Name,
				Version:           artifact.Version,
				Role:              artifact.Role,
				Status:            artifact.Status,
				SizeBytes:         artifact.SizeBytes,
				ChecksumSHA256Hex: artifact.ChecksumSHA256Hex,
				Metadata:          metadata,
			},
			Ref: ref,
			R:   r,
		})
	}
	return loaded, nil
}

func (p *provider) getArtifact(ctx context.Context, options LoadOptions) (*artifactmodel.Artifact, error) {
	var filterOpt []artifactrepo.FilterOption
	if options.ID != nil {
		filterOpt = append(filterOpt, artifactrepo.WithIDs(*options.ID))
	} else if options.ProjectID != nil && options.Name != nil && options.Version != nil {
		filterOpt = append(filterOpt, artifactrepo.WithProjectIDs(*options.ProjectID), artifactrepo.WithVersions(*options.Version), artifactrepo.WithNames(*options.Name))
	} else if options.Ref != nil {
		filterOpt = append(filterOpt, artifactrepo.WithObjectRef(*options.Ref))
	}
	artifact, err := p.artifactRepo.Get(ctx, filterOpt...)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.NewObjectNotFoundError("artifact not found")
		}
		p.logger.Error("failed to get artifact", zap.Error(err))
		return nil, fmt.Errorf("failed to get artifact: %w", err)
	}
	return artifact, nil
}

func (p *provider) listArtifacts(ctx context.Context, appID, projectID *uuid.UUID, ids uuid.UUIDs) ([]*artifactmodel.Artifact, error) {
	var filterOpt []artifactrepo.FilterOption
	if appID != nil {
		filterOpt = append(filterOpt, artifactrepo.WithAppIDs(*appID))
	}
	if projectID != nil {
		filterOpt = append(filterOpt, artifactrepo.WithProjectIDs(*projectID))
	}
	if len(ids) > 0 {
		filterOpt = append(filterOpt, artifactrepo.WithIDs(ids...))
	}
	artifacts, err := p.artifactRepo.UnpaginatedList(ctx, filterOpt...)
	if err != nil {
		p.logger.Error("failed to get artifacts", zap.Error(err))
		return nil, fmt.Errorf("failed to get artifacts: %w", err)
	}
	return artifacts, nil
}

// Validate validates [LoadOptions]. Either ID or ProjectID, Name or Version must be provided.
func (opt LoadOptions) Validate() error {
	return validation.ValidateStruct(&opt,
		validation.Field(&opt.ID, validation.By(validationutil.IsValidUUIDv7), validation.When(
			(opt.ProjectID == nil || opt.Name == nil || opt.Version == nil) && opt.Ref == nil, validation.Required,
		)),
		validation.Field(&opt.ProjectID, validation.By(validationutil.IsValidUUIDv7), validation.When(
			opt.ID == nil && opt.Ref == nil, validation.Required,
		)),
		validation.Field(&opt.Name, validation.Length(1, 128), validation.When(
			opt.ID == nil && opt.Ref == nil, validation.Required,
		)),
		validation.Field(&opt.Version, validation.Length(1, 64), validation.By(validationutil.IsValidSemver), validation.When(
			opt.ID == nil && opt.Ref == nil, validation.Required,
		)),
		validation.Field(&opt.Ref, validation.When(
			opt.ID == nil && (opt.ProjectID == nil || opt.Name == nil || opt.Version == nil), validation.Required,
		)),
	)
}
