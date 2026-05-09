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

package aws_kms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"go.uber.org/zap"
)

const (
	awsKmsProviderType = "aws-kms"
)

type Provider struct {
	client KMSClientAPI
	keyID  string
	logger *zap.Logger
}

func New(ctx context.Context, region, keyID string, logger *zap.Logger) (*Provider, error) {
	var cfgOptions []func(*config.LoadOptions) error
	if region != "" {
		cfgOptions = append(cfgOptions, config.WithRegion(region))
	}
	cfg, err := config.LoadDefaultConfig(ctx, cfgOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Provider{
		client: kms.NewFromConfig(cfg),
		keyID:  keyID,
		logger: logger.With(zap.String("domain", "encryption"), zap.String("provider", awsKmsProviderType)),
	}, nil
}

func (p *Provider) Name() string {
	return awsKmsProviderType
}

func (p *Provider) WrapKey(ctx context.Context, dek []byte) ([]byte, map[string]any, error) {
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

func (p *Provider) UnwrapKey(ctx context.Context, wrapped []byte, metadata map[string]any) ([]byte, error) {
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
