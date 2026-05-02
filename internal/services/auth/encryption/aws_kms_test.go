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

package encryption

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type mockKMSClient struct {
	encryptFunc func(ctx context.Context, params *kms.EncryptInput, optFns ...func(*kms.Options)) (*kms.EncryptOutput, error)
	decryptFunc func(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error)
}

func (m *mockKMSClient) Encrypt(ctx context.Context, params *kms.EncryptInput, optFns ...func(*kms.Options)) (*kms.EncryptOutput, error) {
	return m.encryptFunc(ctx, params, optFns...)
}

func (m *mockKMSClient) Decrypt(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
	return m.decryptFunc(ctx, params, optFns...)
}

func TestAwsKmsProvider_EncryptDecrypt(t *testing.T) {
	t.Parallel()

	logger := zaptest.NewLogger(t)
	keyID := "test-key-id"

	mockClient := &mockKMSClient{
		encryptFunc: func(ctx context.Context, params *kms.EncryptInput, optFns ...func(*kms.Options)) (*kms.EncryptOutput, error) {
			require.Equal(t, keyID, *params.KeyId)
			require.Equal(t, []byte("plaintext-secret"), params.Plaintext)
			return &kms.EncryptOutput{
				CiphertextBlob: []byte("encrypted-secret"),
			}, nil
		},
		decryptFunc: func(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
			require.Equal(t, []byte("encrypted-secret"), params.CiphertextBlob)
			return &kms.DecryptOutput{
				Plaintext: []byte("plaintext-secret"),
			}, nil
		},
	}

	provider := &awsKmsKeyEncryptionProvider{
		client: mockClient,
		keyID:  keyID,
		logger: logger,
	}

	require.Equal(t, awsKmsProviderType, provider.ProviderName())

	// Test Encrypt
	ciphertext, err := provider.Encrypt(context.Background(), []byte("plaintext-secret"))
	require.NoError(t, err)
	expectedCiphertext := base64.StdEncoding.EncodeToString([]byte("encrypted-secret"))
	require.Equal(t, expectedCiphertext, ciphertext)

	// Test Decrypt
	plaintext, err := provider.Decrypt(context.Background(), expectedCiphertext)
	require.NoError(t, err)
	require.Equal(t, []byte("plaintext-secret"), plaintext)
}

func TestAwsKmsProvider_WrapUnwrapKey(t *testing.T) {
	t.Parallel()

	logger := zaptest.NewLogger(t)
	keyID := "test-key-id"

	mockClient := &mockKMSClient{
		encryptFunc: func(ctx context.Context, params *kms.EncryptInput, optFns ...func(*kms.Options)) (*kms.EncryptOutput, error) {
			require.Equal(t, keyID, *params.KeyId)
			require.Equal(t, []byte("dek"), params.Plaintext)
			return &kms.EncryptOutput{CiphertextBlob: []byte("wrapped-dek")}, nil
		},
		decryptFunc: func(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
			require.Equal(t, []byte("wrapped-dek"), params.CiphertextBlob)
			return &kms.DecryptOutput{Plaintext: []byte("dek")}, nil
		},
	}

	provider := &awsKmsKeyEncryptionProvider{
		client: mockClient,
		keyID:  keyID,
		logger: logger,
	}

	wrapped, metadata, err := provider.WrapKey(context.Background(), []byte("dek"))
	require.NoError(t, err)
	require.Equal(t, []byte("wrapped-dek"), wrapped)
	require.Equal(t, keyID, metadata["key_id"])

	dek, err := provider.UnwrapKey(context.Background(), wrapped, metadata)
	require.NoError(t, err)
	require.Equal(t, []byte("dek"), dek)
}
