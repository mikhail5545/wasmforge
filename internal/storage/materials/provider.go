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

package materials

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	entryrepo "github.com/mikhail5545/wasmforge/internal/database/storage/crypto/entry"
	materialrepo "github.com/mikhail5545/wasmforge/internal/database/storage/crypto/material"
	entrymodel "github.com/mikhail5545/wasmforge/internal/models/storage/crypto/entry"
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/storage/crypto/material"
	"github.com/mikhail5545/wasmforge/internal/security/encryption"
	"github.com/mikhail5545/wasmforge/internal/storage/core"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	LoadedMaterials struct {
		AppID     *uuid.UUID
		ProjectID *uuid.UUID
		Materials []LoadedMaterial
	}

	LoadedMaterial struct {
		Material MaterialInfo
		Entries  []MaterialEntry
	}

	MaterialEntry struct {
		ID         uuid.UUID
		MaterialID uuid.UUID
		Info       core.ObjectInfo
		EntryType  entrymodel.CryptoMaterialEntryType
		Reader     io.Reader
	}

	ProviderParams struct {
		DataRoot          string
		EntryRepo         entryrepo.Repository
		MaterialRepo      materialrepo.Repository
		EncryptionService encryption.Service
		RefBuilder        *RefBuilder
		ObjectStore       core.ObjectStore
	}
)

// Provider manages the crypto material loading from the storage provider.
type Provider interface {
	LoadMaterial(ctx context.Context, materialID uuid.UUID) (LoadedMaterial, error)
	LoadMaterials(ctx context.Context, appID, projectID *uuid.UUID) (LoadedMaterials, error)
}

type provider struct {
	dataRoot          string
	objectStore       core.ObjectStore
	refBuilder        *RefBuilder
	entryRepo         entryrepo.Repository
	materialRepo      materialrepo.Repository
	encryptionService encryption.Service
	logger            *zap.Logger
}

func NewProvider(params ProviderParams, logger *zap.Logger) Provider {
	return &provider{
		dataRoot:          params.DataRoot,
		objectStore:       params.ObjectStore,
		refBuilder:        params.RefBuilder,
		entryRepo:         params.EntryRepo,
		materialRepo:      params.MaterialRepo,
		encryptionService: params.EncryptionService,
		logger:            logger.With(zap.String("domain", "storage"), zap.String("component", "material_provider")),
	}
}

func (p *provider) LoadMaterial(ctx context.Context, materialID uuid.UUID) (LoadedMaterial, error) {
	var cleanupFunctions []func()
	defer func() {
		for _, fn := range cleanupFunctions {
			fn()
		}
	}()

	material, err := p.materialRepo.Get(ctx, materialrepo.WithIDs(materialID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return LoadedMaterial{}, core.NewObjectNotFoundError("material not found")
		}
		p.logger.Error("failed to get material", zap.String("material_id", materialID.String()), zap.Error(err))
		return LoadedMaterial{}, fmt.Errorf("failed to load material: %w", err)
	}

	entries, err := p.entryRepo.UnpaginatedList(ctx, entryrepo.WithMaterialIDs(material.ID))
	if err != nil {
		p.logger.Error("failed to list entries", zap.String("material_id", material.ID.String()), zap.Error(err))
		return LoadedMaterial{}, fmt.Errorf("failed to list entries: %w", err)
	}
	if len(entries) == 0 {
		return LoadedMaterial{}, core.NewObjectNotFoundError("no entries found")
	}

	loaded := LoadedMaterial{
		Material: MaterialInfo{
			ID:                 material.ID,
			Name:               material.Name,
			Kind:               material.Kind,
			AppID:              material.AppID,
			ProjectID:          material.ProjectID,
			HasPrivateMaterial: material.HasPrivateMaterial,
			Encrypted:          material.Encrypted,
		},
		Entries: make([]MaterialEntry, 0, len(entries)),
	}

	for _, entry := range entries {
		ref := core.ObjectRef{
			Bucket: core.BucketType(entry.ObjectRefBucket),
			Key:    entry.ObjectRefKey,
		}
		entryRaw, info, err := p.objectStore.Get(ctx, ref)
		if err != nil {
			p.logger.Error("failed to get entry from storage", zap.String("material_id", material.ID.String()), zap.Error(err))
			return LoadedMaterial{}, fmt.Errorf("failed to get entry from storage: %w", err)
		}
		cleanupFunctions = append(cleanupFunctions, func() {
			if err := entryRaw.Close(); err != nil {
				p.logger.Error("failed to close entry", zap.String("material_id", material.ID.String()), zap.Error(err))
			}
		})
		info.Checksum = entry.Checksum

		var data io.Reader = entryRaw
		if material.Encrypted {
			data, err = p.decryptEntry(ctx, entryRaw, entry)
			if err != nil {
				return LoadedMaterial{}, err
			}
		}

		loaded.Entries = append(loaded.Entries, MaterialEntry{
			MaterialID: material.ID,
			Info:       info,
			EntryType:  entry.EntryType,
			Reader:     data,
		})
	}

	return loaded, nil
}

func (p *provider) LoadMaterials(ctx context.Context, appID, projectID *uuid.UUID) (LoadedMaterials, error) {
	if appID == nil && projectID == nil {
		return LoadedMaterials{}, fmt.Errorf("at least one of app_id, project_id must be provided")
	}
	var cleanupFunctions []func()
	defer func() {
		for _, fn := range cleanupFunctions {
			fn()
		}
	}()

	var options []materialrepo.FilterOption
	if appID != nil {
		options = append(options, materialrepo.WithAppIDs(*appID))
	}
	if projectID != nil {
		options = append(options, materialrepo.WithProjectIDs(*projectID))
	}
	materials, err := p.materialRepo.UnpaginatedList(ctx, options...)
	if err != nil {
		p.logger.Error("failed to list materials", zap.Error(err))
		return LoadedMaterials{}, fmt.Errorf("failed to list materials: %w", err)
	}
	if len(materials) == 0 {
		return LoadedMaterials{}, nil
	}
	materialIDs := make([]uuid.UUID, 0, len(materials))
	for _, material := range materials {
		materialIDs = append(materialIDs, material.ID)
	}
	entries, err := p.entryRepo.UnpaginatedList(ctx, entryrepo.WithMaterialIDs(materialIDs...))
	if err != nil {
		p.logger.Error("failed to list entries for materials", zap.Error(err))
		return LoadedMaterials{}, fmt.Errorf("failed to list entries for materials: %w", err)
	}
	if len(entries) == 0 {
		return LoadedMaterials{}, core.NewObjectNotFoundError("no entries found")
	}

	materialIDToEntries := make(map[uuid.UUID][]*entrymodel.CryptoMaterialEntry)
	for i, entry := range entries {
		materialIDToEntries[entry.MaterialID] = append(materialIDToEntries[entry.MaterialID], entries[i])
	}
	materialsMap := make(map[uuid.UUID]*materialmodel.CryptoMaterial)
	for _, material := range materials {
		materialsMap[material.ID] = material
	}
	loadedMaterials := LoadedMaterials{
		AppID:     appID,
		ProjectID: projectID,
		Materials: make([]LoadedMaterial, 0, len(materials)),
	}

	for materialID, entries := range materialIDToEntries {
		material := materialsMap[materialID]
		loaded := LoadedMaterial{
			Material: MaterialInfo{
				ID:                 material.ID,
				Name:               material.Name,
				Kind:               material.Kind,
				AppID:              material.AppID,
				ProjectID:          material.ProjectID,
				HasPrivateMaterial: material.HasPrivateMaterial,
				Encrypted:          material.Encrypted,
			},
			Entries: make([]MaterialEntry, 0, len(entries)),
		}

		for _, entry := range entries {
			ref := core.ObjectRef{
				Bucket: core.BucketType(entry.ObjectRefBucket),
				Key:    entry.ObjectRefKey,
			}
			entryRaw, info, err := p.objectStore.Get(ctx, ref)
			if err != nil {
				p.logger.Error("failed to get entry from storage", zap.String("material_id", entry.MaterialID.String()), zap.Error(err))
				return LoadedMaterials{}, fmt.Errorf("failed to get entry from storage: %w", err)
			}
			cleanupFunctions = append(cleanupFunctions, func() {
				if err := entryRaw.Close(); err != nil {
					p.logger.Error("failed to close entry", zap.String("material_id", entry.MaterialID.String()), zap.Error(err))
				}
			})
			info.Checksum = entry.Checksum

			var data io.Reader = entryRaw
			if material.Encrypted {
				data, err = p.decryptEntry(ctx, entryRaw, entry)
				if err != nil {
					return LoadedMaterials{}, err
				}
			}

			loaded.Entries = append(loaded.Entries, MaterialEntry{
				MaterialID: material.ID,
				Info:       info,
				EntryType:  entry.EntryType,
				Reader:     data,
			})
		}
		loadedMaterials.Materials = append(loadedMaterials.Materials, loaded)
	}
	return loadedMaterials, nil
}

func (p *provider) decryptEntry(ctx context.Context, enc io.ReadCloser, entry *entrymodel.CryptoMaterialEntry) (io.Reader, error) {
	raw, err := io.ReadAll(enc)
	if err != nil {
		p.logger.Error("failed to decrypt entry", zap.String("material_id", entry.ID.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to decrypt entry: %w", err)
	}
	metadata := make(map[string]any)
	if entry.EncryptionProviderMetadata != "" {
		if err := json.Unmarshal([]byte(entry.EncryptionProviderMetadata), &metadata); err != nil {
			p.logger.Error("failed to unmarshal metadata", zap.String("material_id", entry.ID.String()), zap.Error(err))
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	envelope := encryption.Envelope{
		Ciphertext:       raw,
		WrappedDEK:       []byte(*entry.WrappedDEK),
		Nonce:            []byte(*entry.EncryptionNonce),
		Algorithm:        *entry.EncryptionAlgorithm,
		Provider:         *entry.EncryptionProvider,
		ProviderMetadata: metadata,
	}
	dek, err := p.encryptionService.Decrypt(ctx, envelope)
	if err != nil {
		p.logger.Error("failed to decrypt entry", zap.String("material_id", entry.ID.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to decrypt entry: %w", err)
	}
	return bytes.NewReader(dek), nil
}
