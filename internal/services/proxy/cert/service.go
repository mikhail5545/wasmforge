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

package cert

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime/multipart"

	"github.com/mikhail5545/wasmforge/internal/crypto"
	configrepo "github.com/mikhail5545/wasmforge/internal/database/proxy/config"
	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	configmodel "github.com/mikhail5545/wasmforge/internal/models/proxy/config"
	"github.com/mikhail5545/wasmforge/internal/proxy/server"
	"github.com/mikhail5545/wasmforge/internal/uploads"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	server        *server.Server
	configRepo    configrepo.Repository
	uploadManager uploads.Manager
	logger        *zap.Logger
}

func New(server *server.Server, configRepo configrepo.Repository, uploadManager uploads.Manager, logger *zap.Logger) *Service {
	return &Service{
		server:        server,
		configRepo:    configRepo,
		uploadManager: uploadManager,
		logger:        logger.With(zap.String("component", "proxy_cert_service")),
	}
}

func (s *Service) UploadCerts(ctx context.Context, certFile, keyFile *multipart.FileHeader) error {
	if certFile == nil || keyFile == nil {
		return fmt.Errorf("cert file or key file is nil")
	}

	return s.configRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.configRepo.WithTx(tx)

		s.logger.Debug("uploading TLS certs", zap.String("cert_file", certFile.Filename), zap.String("key_file", keyFile.Filename))

		certHash, err := s.uploadManager.FromMultipartFile(certFile, certFile.Filename, uploads.CertUpload)
		if err != nil {
			tx.Rollback()
			s.logger.Error("failed to upload cert file", zap.Error(err))
			return err
		}
		keyHash, err := s.uploadManager.FromMultipartFile(keyFile, keyFile.Filename, uploads.CertUpload)
		if err != nil {
			tx.Rollback()
			s.logger.Error("failed to upload key file", zap.Error(err))
			return err
		}

		s.logger.Info("successfully uploaded TLS certs", zap.String("cert_hash", certHash), zap.String("key_hash", keyHash))

		if err := txRepo.Updates(ctx, map[string]any{
			"tls_cert_path": certFile.Filename,
			"tls_cert_hash": certHash,
			"tls_key_path":  keyFile.Filename,
			"tls_key_hash":  keyHash,
			"tls_enabled":   true,
		}); err != nil {
			s.logger.Error("failed to update proxy config with new cert paths and hashes", zap.Error(err))
			return fmt.Errorf("failed to update proxy config with new cert paths and hashes: %w", err)
		}
		return nil
	})
}

func (s *Service) LoadCerts(config *configmodel.Config) (*tls.Config, error) {
	if !config.TLSEnabled {
		return nil, inerrors.NewConflictError("tls is not enabled in proxy config, cannot load certs")
	}
	if config.TLSCertPath == nil || config.TLSKeyPath == nil {
		s.logger.Warn("TLS credentials are enabled but cert path or key path is nil, cannot load certs")
		return nil, fmt.Errorf("unable to load TLS certs: cert path or key path is nil in proxy config")
	}

	certData, err := s.uploadManager.Read(*config.TLSCertPath, uploads.CertUpload)
	if err != nil {
		s.logger.Error("failed to read cert file from storage", zap.Error(err))
		return nil, fmt.Errorf("failed to read cert file from storage: %w", err)
	}
	keyData, err := s.uploadManager.Read(*config.TLSKeyPath, uploads.CertUpload)
	if err != nil {
		s.logger.Error("failed to read key file from storage", zap.Error(err))
		return nil, fmt.Errorf("failed to read key file from storage: %w", err)
	}

	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		s.logger.Error("failed to parse TLS cert and key data", zap.Error(err))
		return nil, fmt.Errorf("failed to parse TLS cert and key data: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

func (s *Service) RemoveCerts(ctx context.Context) error {
	return s.configRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.configRepo.WithTx(tx)

		s.logger.Debug("removing TLS certs")

		config, err := txRepo.Get(ctx)
		if err != nil {
			s.logger.Error("failed to get proxy config for cert removal", zap.Error(err))
			return fmt.Errorf("failed to get proxy config for cert removal: %w", err)
		}

		if err := s.deleteCertFiles(config.TLSCertPath, config.TLSKeyPath); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete cert files: %w", err)
		}

		if err := txRepo.Updates(ctx, map[string]any{
			"tls_cert_path": nil,
			"tls_cert_hash": nil,
			"tls_key_path":  nil,
			"tls_key_hash":  nil,
			"tls_enabled":   false,
		}); err != nil {
			s.logger.Error("failed to update proxy config to disable TLS and clear cert paths and hashes", zap.Error(err))
			return fmt.Errorf("failed to update proxy config to disable TLS and clear cert paths and hashes: %w", err)
		}
		s.logger.Info("successfully removed TLS certs and updated proxy config")
		return nil
	})
}

func (s *Service) GenerateSelfSignedCerts(ctx context.Context, req *configmodel.GenerateCertificatesRequest) error {
	if err := req.Validate(); err != nil {
		return inerrors.NewValidationError(err)
	}

	s.logger.Debug("generating self-signed TLS certs")

	certPem, keyPem, err := crypto.GenerateSlefSignedCerts(req.CommonName, req.ValidDays, req.RsaBits, s.logger)
	if err != nil {
		s.logger.Error("failed to generate self-signed TLS certs", zap.Error(err))
		return fmt.Errorf("failed to generate self-signed TLS certs: %w", err)
	}

	s.logger.Debug("successfully generated self-signed TLS certs, uploading to storage")
	certHash, err := s.uploadManager.FromBytes(certPem, "selfsigned_cert.pem", uploads.CertUpload)
	if err != nil {
		s.logger.Error("failed to upload generated cert PEM data", zap.Error(err))
		return fmt.Errorf("failed to upload generated cert PEM data: %w", err)
	}
	keyHash, err := s.uploadManager.FromBytes(keyPem, "selfsigned_key.pem", uploads.CertUpload)
	if err != nil {
		s.logger.Error("failed to upload generated key PEM data", zap.Error(err))
		return fmt.Errorf("failed to upload generated key PEM data: %w", err)
	}

	s.logger.Info("successfully generated and uploaded self-signed TLS certs", zap.String("cert_hash", certHash), zap.String("key_hash", keyHash))

	return s.configRepo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.configRepo.WithTx(tx)

		if err := txRepo.Updates(ctx, map[string]any{
			"tls_cert_path": "selfsigned_cert.pem",
			"tls_cert_hash": certHash,
			"tls_key_path":  "selfsigned_key.pem",
			"tls_key_hash":  keyHash,
			"tls_enabled":   true,
		}); err != nil {
			s.logger.Error("failed to update proxy config with new self-signed cert paths and hashes", zap.Error(err))
			return fmt.Errorf("failed to update proxy config with new self-signed cert paths and hashes: %w", err)
		}
		return nil
	})
}
