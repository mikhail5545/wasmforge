/*
 * Copyright (c) 2024-2026. Mikhail Kulik.
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

package encryption

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"go.uber.org/zap"
)

const (
	awsKmsProviderType = "aws-kms"
)

// KMSClientAPI defines the interface for the KMS methods we need, allowing for mocking in tests.
type KMSClientAPI interface {
	Encrypt(ctx context.Context, params *kms.EncryptInput, optFns ...func(*kms.Options)) (*kms.EncryptOutput, error)
	Decrypt(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error)
}

type awsKmsKeyEncryptionProvider struct {
	client KMSClientAPI
	keyID  string
	logger *zap.Logger
}

func NewAwsKmsKeyEncryptionProvider(ctx context.Context, region, keyID string, logger *zap.Logger) (KeyEncryptionProvider, error) {
	var cfgOptions []func(*config.LoadOptions) error
	if region != "" {
		cfgOptions = append(cfgOptions, config.WithRegion(region))
	}
	cfg, err := config.LoadDefaultConfig(ctx, cfgOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &awsKmsKeyEncryptionProvider{
		client: kms.NewFromConfig(cfg),
		keyID:  keyID,
		logger: logger.With(zap.String("provider", awsKmsProviderType)),
	}, nil
}

func (p *awsKmsKeyEncryptionProvider) ProviderName() string {
	return awsKmsProviderType
}

func (p *awsKmsKeyEncryptionProvider) WrapKey(ctx context.Context, dek []byte) ([]byte, map[string]any, error) {
	res, err := p.client.Encrypt(ctx, &kms.EncryptInput{
		KeyId:     aws.String(p.keyID),
		Plaintext: dek,
	})
	if err != nil {
		p.logger.Error("failed to encrypt with AWS KMS", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to encrypt with AWS KMS: %w", err)
	}
	return res.CiphertextBlob, map[string]any{"key_id": p.keyID}, nil
}

func (p *awsKmsKeyEncryptionProvider) UnwrapKey(ctx context.Context, wrapped []byte, metadata map[string]any) ([]byte, error) {
	_ = metadata
	res, err := p.client.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob: wrapped,
	})
	if err != nil {
		p.logger.Error("failed to decrypt with AWS KMS", zap.Error(err))
		return nil, fmt.Errorf("failed to decrypt with AWS KMS: %w", err)
	}
	return res.Plaintext, nil
}

func (p *awsKmsKeyEncryptionProvider) Encrypt(ctx context.Context, plaintextKey []byte) (string, error) {
	wrapped, _, err := p.WrapKey(ctx, plaintextKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(wrapped), nil
}

func (p *awsKmsKeyEncryptionProvider) Decrypt(ctx context.Context, ciphertextKey string) ([]byte, error) {
	blob, err := base64.StdEncoding.DecodeString(ciphertextKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}
	return p.UnwrapKey(ctx, blob, nil)
}
