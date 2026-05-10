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
	"fmt"
	"io"
	"mime/multipart"
	"os"

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
	UploadFromMultipartParams struct {
		UploadName string
		ProjectID  uuid.UUID
		AppID      *uuid.UUID
		Parts      []UploadFromMultipartPart
	}

	UploadFromMultipartPart struct {
		Name      string
		File      *multipart.FileHeader
		SizeBytes int64
		Metadata  map[string]any
	}

	UploadFromPathParams struct {
		UploadName string
		ProjectID  uuid.UUID
		AppID      *uuid.UUID
		Parts      []UploadFromPathPart
	}

	UploadFromPathPart struct {
		Name      string
		Source    string
		SizeBytes int64
		Metadata  map[string]any
	}

	UploadFromBytesParams struct {
		UploadName string
		ProjectID  uuid.UUID
		AppID      *uuid.UUID
		Parts      []UploadFromBytesPart
	}

	UploadFromBytesPart struct {
		Name      string
		Data      []byte
		SizeBytes int64
		Metadata  map[string]any
	}

	UploadParams struct {
		UploadName string
		ProjectID  uuid.UUID
		AppID      *uuid.UUID
		Parts      []UploadPart
	}

	UploadPart struct {
		Name      string
		Reader    io.Reader
		SizeBytes int64
		Metadata  map[string]any
	}

	Config struct {
		DataRoot           string
		MaxUploadSizeBytes int64
	}

	MaterialInfo struct {
		ID                 uuid.UUID
		Name               string
		Kind               materialmodel.CryptoMaterialKind
		AppID              *uuid.UUID
		ProjectID          uuid.UUID
		Entries            int
		HasPrivateMaterial bool
		Encrypted          bool
	}

	ManagerParams struct {
		MaterialRepo materialrepo.Repository
		EntryRepo    entryrepo.Repository
		ObjectStore  core.ObjectStore
	}
)

// Manager manages the whole process of crypto material uploads and validation.
// It provides multiple entry points for different upload sources, such as multipart form data, raw bytes, or file paths,
// but all of them will be converted into a common format with readers and passed to the main Upload method.
// Currently supported uploads are:
//
// - Public cert: single cert file, unencrypted
// - Key Pair: single CA pair (matching certificate + private key) or 1 private and 1 public key (total 2 files, encrypted)
// - CA Bundle: multiple certificates, no keys, all CAs, no private material, unencrypted
// - Trust Bundle: multiple certificates, no keys, not all CAs, no private material, unencrypted
type Manager interface {
	// UploadFromMultipart wraps Upload and creates readers from provided [UploadFromMultipartPart.File] entries.
	UploadFromMultipart(ctx context.Context, params UploadFromMultipartParams) (MaterialInfo, error)
	// UploadFromBytes wraps Upload and creates readers from provided [UploadFromBytesPart.Data] entries.
	UploadFromBytes(ctx context.Context, params UploadFromBytesParams) (MaterialInfo, error)
	// UploadFromPath wraps Upload and creates readers from provided [UploadFromPathPart.Source] entries.
	UploadFromPath(ctx context.Context, params UploadFromPathParams) (MaterialInfo, error)
	// Upload manages the whole process of the certificates upload. It can handle the upload of
	// multiple files for different kinds of crypto material record. For example, [materialmodel.CryptoMaterialKindPublicCert] (1 cert file, unencrypted),
	// [materialmodel.CryptoMaterialKindKeyPair] (1 cert file + 1 private key file, encrypted), [materialmodel.CryptoMaterialKindTrustBundle]
	// (multiple cert files, unencrypted), etc.
	//
	// For multiple input files, the caller should provide a list of files with metadata in the [UploadParams.Parts].
	// Each file will be saved as a temp file and validated by the [Validator].
	//
	// Upload cycle: temporary file(s) -> validation -> move file(s) to permanent location -> determinicstic processing (db entries)
	Upload(ctx context.Context, params UploadParams) (MaterialInfo, error)
}

type manager struct {
	cfg               Config
	objectStore       core.ObjectStore
	refBuilder        *RefBuilder
	entryRepo         entryrepo.Repository
	materialRepo      materialrepo.Repository
	encryptionService encryption.Service
	validator         *Validator
	logger            *zap.Logger
}

func NewManager(cfg Config, params ManagerParams, logger *zap.Logger) Manager {
	return &manager{
		objectStore:  params.ObjectStore,
		materialRepo: params.MaterialRepo,
		entryRepo:    params.EntryRepo,
		validator:    NewValidator(logger),
		refBuilder:   NewRefBuilder(cfg.DataRoot),
		logger:       logger.With(zap.String("domain", "storage"), zap.String("component", "material_manager")),
	}
}

// UploadFromMultipart wraps Upload and creates readers from provided [UploadFromMultipartPart.File] entries.
func (m *manager) UploadFromMultipart(ctx context.Context, params UploadFromMultipartParams) (MaterialInfo, error) {
	var cleanupFunctions []func()
	defer func() {
		for _, fn := range cleanupFunctions {
			fn()
		}
	}()

	uploadParams := UploadParams{
		UploadName: params.UploadName,
		ProjectID:  params.ProjectID,
		AppID:      params.AppID,
		Parts:      make([]UploadPart, 0, len(params.Parts)),
	}

	for _, part := range params.Parts {
		r, err := part.File.Open()
		if err != nil {
			return MaterialInfo{}, fmt.Errorf("failed to open multipart file: %w", err)
		}
		cleanupFunctions = append(cleanupFunctions, func() {
			if err := r.Close(); err != nil {
				m.logger.Error("failed to close multipart file", zap.Error(err))
			}
		})

		uploadParams.Parts = append(uploadParams.Parts, UploadPart{
			Name:      part.Name,
			Reader:    r,
			SizeBytes: part.SizeBytes,
			Metadata:  part.Metadata,
		})
	}
	return m.Upload(ctx, uploadParams)
}

// UploadFromBytes wraps Upload and creates readers from provided [UploadFromBytesPart.Data] entries.
func (m *manager) UploadFromBytes(ctx context.Context, params UploadFromBytesParams) (MaterialInfo, error) {
	uploadParams := UploadParams{
		UploadName: params.UploadName,
		ProjectID:  params.ProjectID,
		AppID:      params.AppID,
		Parts:      make([]UploadPart, 0, len(params.Parts)),
	}

	for _, part := range params.Parts {
		uploadParams.Parts = append(uploadParams.Parts, UploadPart{
			Name:      part.Name,
			Reader:    bytes.NewReader(part.Data),
			SizeBytes: part.SizeBytes,
		})
	}
	return m.Upload(ctx, uploadParams)
}

// UploadFromPath wraps Upload and creates readers from provided [UploadFromPathPart.Source] entries.
func (m *manager) UploadFromPath(ctx context.Context, params UploadFromPathParams) (MaterialInfo, error) {
	var cleanupFunctions []func()
	defer func() {
		for _, fn := range cleanupFunctions {
			fn()
		}
	}()

	uploadParams := UploadParams{
		UploadName: params.UploadName,
		ProjectID:  params.ProjectID,
		AppID:      params.AppID,
		Parts:      make([]UploadPart, 0, len(params.Parts)),
	}

	for _, part := range params.Parts {
		file, err := os.Open(part.Source)
		if err != nil {
			return MaterialInfo{}, fmt.Errorf("failed to open file %s: %w", part.Source, err)
		}
		cleanupFunctions = append(cleanupFunctions, func() {
			if err := file.Close(); err != nil {
				m.logger.Error("failed to close file after upload", zap.Error(err))
			}
		})

		uploadParams.Parts = append(uploadParams.Parts, UploadPart{
			Name:      part.Name,
			Reader:    file,
			SizeBytes: part.SizeBytes,
			Metadata:  part.Metadata,
		})
	}
	return m.Upload(ctx, uploadParams)
}

// Upload manages the whole process of the crypto material upload. It can handle the upload of
// multiple files for different kinds of crypto material record. For example, [materialmodel.CryptoMaterialKindPublicCert] (1 cert file, unencrypted),
// [materialmodel.CryptoMaterialKindKeyPair] (1 cert file + 1 private key file, encrypted), [materialmodel.CryptoMaterialKindTrustBundle]
// (multiple cert files, unencrypted), etc.
//
// For multiple input files, the caller should provide a list of files with metadata in the [UploadParams.Parts].
// Each file will be saved as a temp file and validated by the [Validator].
//
// Upload cycle: temporary file(s) -> validation -> move file(s) to permanent location -> determinicstic processing (db entries)
func (m *manager) Upload(ctx context.Context, params UploadParams) (MaterialInfo, error) {
	var cleanupFuncs []func()
	defer func() {
		for _, fn := range cleanupFuncs {
			fn()
		}
	}()

	m.logger.Debug("starting certificate upload", zap.String("upload_name", params.UploadName), zap.Int("to_upload_parts", len(params.Parts)))

	var validationParts []ValidationPart
	metadata := make(map[core.ObjectRef]string)

	for i, part := range params.Parts {
		if part.SizeBytes > m.cfg.MaxUploadSizeBytes {
			return MaterialInfo{}, core.NewSizeLimitExceededError(m.cfg.MaxUploadSizeBytes)
		}

		tempRef, err := m.refBuilder.BuildTemp()
		if err != nil {
			return MaterialInfo{}, err
		}
		metadata[tempRef] = part.Name

		dest, err := os.Create(tempRef.Key)
		if err != nil {
			m.logger.Error("failed to create temp file", zap.Error(err))
			return MaterialInfo{}, fmt.Errorf("failed to create temp file: %w", err)
		}

		cleanupFuncs = append(cleanupFuncs, func() {
			if err := os.Remove(tempRef.Key); err != nil {
				m.logger.Error("failed to remove temp file", zap.Error(err))
			}
		})

		_, err = io.Copy(dest, part.Reader)
		if err != nil {
			if err := dest.Close(); err != nil {
				m.logger.Error("failed to close temp file", zap.Error(err))
			}
			return MaterialInfo{}, fmt.Errorf("failed to write part to temp file: %w", err)
		}
		m.logger.Debug("saved file to temporary storage", zap.String("key", tempRef.Key))
		if err := dest.Close(); err != nil {
			m.logger.Error("failed to close temp file", zap.Error(err))
			return MaterialInfo{}, fmt.Errorf("failed to close temp file: %w", err)
		}

		// Re-open for reading
		readDest, err := os.Open(tempRef.Key)
		if err != nil {
			return MaterialInfo{}, fmt.Errorf("failed to re-open temp file for validation: %w", err)
		}

		cleanupFuncs = append(cleanupFuncs, func() {
			if err := readDest.Close(); err != nil {
				m.logger.Error("failed to close temp file", zap.Error(err))
			}
		})

		partName := part.Name
		if partName == "" {
			partName = fmt.Sprintf("part-%d", i)
		}

		meta, err := json.Marshal(part.Metadata)
		if err != nil {
			m.logger.Error("failed to marshal metadata", zap.Error(err))
			return MaterialInfo{}, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		validationParts = append(validationParts, ValidationPart{
			Name:      partName,
			Reader:    readDest,
			SizeHint:  part.SizeBytes,
			ObjectRef: tempRef,
			Metadata:  string(meta),
		})
	}

	validationRes, err := m.validator.Validate(ValidationInput{
		Parts: validationParts,
	})
	if err != nil {
		return MaterialInfo{}, err
	}
	m.logger.Debug("validation completed, proceeding to permanent save", zap.String("upload_name", params.UploadName), zap.String("upload_kind", validationRes.Kind.String()))

	switch validationRes.Kind {
	case materialmodel.CryptoMaterialKindPublicCert:
		return m.savePublicCertMaterial(ctx, validationRes, params)
	case materialmodel.CryptoMaterialKindKeyPair:
		return m.saveKeyPairMaterial(ctx, validationRes, params)
	case materialmodel.CryptoMaterialKindTrustBundle:
		return m.saveTrustBundleMaterial(ctx, validationRes, params)
	case materialmodel.CryptoMaterialKindCABundle:
		return m.saveCABundleMaterial(ctx, validationRes, params)
	default:
		return MaterialInfo{}, fmt.Errorf("unsupported certificate kind: %s", validationRes.Kind)
	}
}

func (m *manager) savePublicCertMaterial(
	ctx context.Context,
	result *ValidationResult,
	params UploadParams,
) (MaterialInfo, error) {
	// When validation result classified as public cert kind, we expect exactly ONE certificate in the result,
	// and we will create a single material with one entry for that cert.
	// Validator should ensure that these assumptions are met, so we will panic if not, since it indicates a bug in the validator.
	if len(result.Certs) != 1 {
		panic(fmt.Sprintf("expected exactly one cert in validation result for public cert kind, got %d", len(result.Certs)))
	}

	materialID, err := uuid.NewV7()
	if err != nil {
		return MaterialInfo{}, fmt.Errorf("failed to generate material ID: %w", err)
	}

	var material *materialmodel.CryptoMaterial
	var entry *entrymodel.CryptoMaterialEntry
	var savedRef core.ObjectRef

	err = m.materialRepo.DB().Transaction(func(tx *gorm.DB) error {
		txMaterialRepo := m.materialRepo.WithTx(tx)
		txEntryRepo := m.entryRepo.WithTx(tx)

		cert := result.Certs[0]

		material = &materialmodel.CryptoMaterial{
			ID:                 materialID,
			ProjectID:          params.ProjectID,
			AppID:              params.AppID,
			Kind:               materialmodel.CryptoMaterialKindPublicCert,
			Name:               cert.PartName,
			HasPrivateMaterial: false,
			Encrypted:          false,
		}

		entry, err = m.saveEntry(ctx, materialID, params, entrymodel.CryptoMaterialEntryTypeCertificate, 0, cert.ObjectRef, &cert, false)
		if err != nil {
			return err
		}

		savedRef = m.refBuilder.Build(BuildParams{
			ObjectID:  entry.ID,
			ProjectID: params.ProjectID,
			AppID:     params.AppID,
			Encrypted: false,
			Extension: ".crt",
		})

		if err := txMaterialRepo.Create(ctx, material); err != nil {
			return fmt.Errorf("failed to save crypto material: %w", err)
		}
		if err := txEntryRepo.Create(ctx, entry); err != nil {
			return fmt.Errorf("failed to save crypto material entry: %w", err)
		}

		return nil
	})

	if err != nil {
		if savedRef.Key != "" {
			_ = m.objectStore.Delete(ctx, savedRef)
		}
		return MaterialInfo{}, err
	}

	return MaterialInfo{
		ID:                 material.ID,
		Name:               material.Name,
		AppID:              material.AppID,
		ProjectID:          material.ProjectID,
		Encrypted:          material.Encrypted,
		HasPrivateMaterial: material.HasPrivateMaterial,
		Kind:               materialmodel.CryptoMaterialKindPublicCert,
		Entries:            1,
	}, nil
}

func (m *manager) saveKeyPairMaterial(ctx context.Context, result *ValidationResult, params UploadParams) (MaterialInfo, error) {
	materialID, err := uuid.NewV7()
	if err != nil {
		return MaterialInfo{}, fmt.Errorf("failed to generate material ID: %w", err)
	}

	var material *materialmodel.CryptoMaterial
	var entries []*entrymodel.CryptoMaterialEntry
	var savedRefs []core.ObjectRef

	err = m.materialRepo.DB().Transaction(func(tx *gorm.DB) error {
		txMaterialRepo := m.materialRepo.WithTx(tx)
		txEntryRepo := m.entryRepo.WithTx(tx)

		material = &materialmodel.CryptoMaterial{
			ID:                 materialID,
			ProjectID:          params.ProjectID,
			AppID:              params.AppID,
			Kind:               materialmodel.CryptoMaterialKindKeyPair,
			Name:               params.UploadName,
			HasPrivateMaterial: true,
			Encrypted:          true,
		}

		// Always save the private key first (at position 0)
		pk := result.PrivateKeys[0]
		pkEntry, err := m.saveEntry(ctx, materialID, params, entrymodel.CryptoMaterialEntryTypePrivateKey, 0, pk.ObjectRef, &pk, true)
		if err != nil {
			return err
		}
		entries = append(entries, pkEntry)
		savedRefs = append(savedRefs, m.refBuilder.Build(BuildParams{
			ObjectID:  pkEntry.ID,
			ProjectID: params.ProjectID,
			AppID:     params.AppID,
			Encrypted: true,
			Extension: ".key",
		}))

		// Then save the matching cert or public key (at position 1)
		if len(result.CAPairs) == 1 {
			cert := result.CAPairs[0].Cert
			certEntry, err := m.saveEntry(ctx, materialID, params, entrymodel.CryptoMaterialEntryTypeCertificate, 1, cert.ObjectRef, &cert, true)
			if err != nil {
				return err
			}
			entries = append(entries, certEntry)
			savedRefs = append(savedRefs, m.refBuilder.Build(BuildParams{
				ObjectID:  certEntry.ID,
				ProjectID: params.ProjectID,
				AppID:     params.AppID,
				Encrypted: true,
				Extension: ".crt",
			}))
		} else if len(result.PublicKeys) == 1 {
			pub := result.PublicKeys[0]
			pubEntry, err := m.saveEntry(ctx, materialID, params, entrymodel.CryptoMaterialEntryTypePublicKey, 1, pub.ObjectRef, &pub, false)
			if err != nil {
				return err
			}
			entries = append(entries, pubEntry)
			savedRefs = append(savedRefs, m.refBuilder.Build(BuildParams{
				ObjectID:  pubEntry.ID,
				ProjectID: params.ProjectID,
				AppID:     params.AppID,
				Encrypted: false,
				Extension: ".key",
			}))
		}

		if err := txMaterialRepo.Create(ctx, material); err != nil {
			return fmt.Errorf("failed to create material: %w", err)
		}
		for _, e := range entries {
			if err := txEntryRepo.Create(ctx, e); err != nil {
				return fmt.Errorf("failed to create entry: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		for _, ref := range savedRefs {
			_ = m.objectStore.Delete(ctx, ref)
		}
		return MaterialInfo{}, err
	}

	return MaterialInfo{
		ID:                 material.ID,
		Name:               material.Name,
		Kind:               material.Kind,
		AppID:              material.AppID,
		ProjectID:          material.ProjectID,
		Entries:            len(entries),
		HasPrivateMaterial: material.HasPrivateMaterial,
		Encrypted:          material.Encrypted,
	}, nil
}

func (m *manager) saveTrustBundleMaterial(ctx context.Context, result *ValidationResult, params UploadParams) (MaterialInfo, error) {
	return m.saveBundleMaterial(ctx, result, params, materialmodel.CryptoMaterialKindTrustBundle)
}

func (m *manager) saveCABundleMaterial(ctx context.Context, result *ValidationResult, params UploadParams) (MaterialInfo, error) {
	return m.saveBundleMaterial(ctx, result, params, materialmodel.CryptoMaterialKindCABundle)
}

func (m *manager) saveBundleMaterial(ctx context.Context, result *ValidationResult, params UploadParams, kind materialmodel.CryptoMaterialKind) (MaterialInfo, error) {
	materialID, err := uuid.NewV7()
	if err != nil {
		return MaterialInfo{}, fmt.Errorf("failed to generate material ID: %w", err)
	}

	var material *materialmodel.CryptoMaterial
	var entries []*entrymodel.CryptoMaterialEntry
	var savedRefs []core.ObjectRef

	err = m.materialRepo.DB().Transaction(func(tx *gorm.DB) error {
		txMaterialRepo := m.materialRepo.WithTx(tx)
		txEntryRepo := m.entryRepo.WithTx(tx)

		material = &materialmodel.CryptoMaterial{
			ID:                 materialID,
			ProjectID:          params.ProjectID,
			AppID:              params.AppID,
			Kind:               kind,
			Name:               params.UploadName,
			HasPrivateMaterial: false,
			Encrypted:          false,
		}

		for i, cert := range result.Certs {
			certEntry, err := m.saveEntry(ctx, materialID, params, entrymodel.CryptoMaterialEntryTypeCertificate, i, cert.ObjectRef, &cert, false)
			if err != nil {
				return err
			}
			entries = append(entries, certEntry)
			savedRefs = append(savedRefs, m.refBuilder.Build(BuildParams{
				ObjectID:  certEntry.ID,
				ProjectID: params.ProjectID,
				AppID:     params.AppID,
				Encrypted: false,
				Extension: ".crt",
			}))
		}

		if err := txMaterialRepo.Create(ctx, material); err != nil {
			return fmt.Errorf("failed to create material: %w", err)
		}

		if err := txEntryRepo.Create(ctx, entries...); err != nil {
			m.logger.Error("failed to save crypto material entries", zap.Error(err))
			return fmt.Errorf("failed to create entries: %w", err)
		}

		return nil
	})

	if err != nil {
		for _, ref := range savedRefs {
			_ = m.objectStore.Delete(ctx, ref)
		}
		return MaterialInfo{}, err
	}

	return MaterialInfo{
		ID:                 material.ID,
		Name:               material.Name,
		Kind:               material.Kind,
		AppID:              material.AppID,
		ProjectID:          material.ProjectID,
		Entries:            len(entries),
		HasPrivateMaterial: material.HasPrivateMaterial,
		Encrypted:          material.Encrypted,
	}, nil
}

func (m *manager) saveEntry(
	ctx context.Context,
	materialID uuid.UUID,
	params UploadParams,
	entryType entrymodel.CryptoMaterialEntryType,
	position int,
	tempRef core.ObjectRef,
	info any,
	encrypt bool,
) (*entrymodel.CryptoMaterialEntry, error) {
	entryID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate entry ID: %w", err)
	}

	file, err := os.Open(tempRef.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to open temp file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			m.logger.Error("failed to close temp file", zap.Error(err))
		}
	}()

	var ext string
	switch entryType {
	case entrymodel.CryptoMaterialEntryTypeCertificate:
		ext = ".crt"
	case entrymodel.CryptoMaterialEntryTypePrivateKey, entrymodel.CryptoMaterialEntryTypePublicKey:
		ext = ".key"
	}
	ref := m.refBuilder.Build(BuildParams{
		ObjectID:  entryID,
		ProjectID: params.ProjectID,
		AppID:     params.AppID,
		Encrypted: encrypt,
		Extension: ext,
	})

	var objectInfo core.ObjectInfo
	var envelope *encryption.Envelope

	if encrypt {
		data, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read temp file: %w", err)
		}
		env, err := m.encryptionService.Encrypt(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt data: %w", err)
		}
		envelope = &env
		objectInfo, err = m.objectStore.Put(ctx, ref, bytes.NewReader(env.Ciphertext))
		if err != nil {
			return nil, fmt.Errorf("failed to save encrypted object: %w", err)
		}
	} else {
		objectInfo, err = m.objectStore.Put(ctx, ref, file)
		if err != nil {
			return nil, fmt.Errorf("failed to save object: %w", err)
		}
	}

	entry := &entrymodel.CryptoMaterialEntry{
		ID:              entryID,
		MaterialID:      materialID,
		EntryType:       entryType,
		Position:        position,
		ObjectRefBucket: string(ref.Bucket),
		ObjectRefKey:    ref.Key,
		Checksum:        objectInfo.Checksum,
		SizeBytes:       objectInfo.SizeBytes,
	}

	if envelope != nil {
		wrappedDEK := string(envelope.WrappedDEK)
		nonce := string(envelope.Nonce)
		entry.WrappedDEK = &wrappedDEK
		entry.EncryptionNonce = &nonce
		entry.EncryptionAlgorithm = &envelope.Algorithm
		entry.EncryptionProvider = &envelope.Provider

		if len(envelope.ProviderMetadata) > 0 {
			metaBytes, err := json.Marshal(envelope.ProviderMetadata)
			if err != nil {
				m.logger.Error("failed to serialize provider metadata", zap.Error(err))
				return nil, fmt.Errorf("failed to marshal provider metadata: %w", err)
			}
			metaStr := string(metaBytes)
			entry.EncryptionProviderMetadata = metaStr
		}
	}

	switch v := info.(type) {
	case *CertInfo:
		entry.FingerprintSHA256Hex = v.CertSHA256Hex
		entry.Algorithm = &v.PublicKey.Algorithm
		entry.Details = &v.PublicKey.Details
		entry.Subject = &v.Subject
		entry.Issuer = &v.Issuer
		entry.SerialHex = &v.SerialHex
		entry.NotBefore = &v.NotBefore
		entry.NotAfter = &v.NotAfter
		entry.IsCA = v.IsCA
		entry.MetadataJSON = v.Metadata
	case *PrivateKeyInfo:
		entry.FingerprintSHA256Hex = v.PublicKey.Fingerprint
		entry.Algorithm = &v.Algorithm
		entry.Details = &v.Details
		entry.MetadataJSON = v.Metadata
	case *PublicKeyInfo:
		entry.FingerprintSHA256Hex = v.Fingerprint
		entry.Algorithm = &v.Algorithm
		entry.Details = &v.Details
		entry.MetadataJSON = v.Metadata
	}

	return entry, nil
}
