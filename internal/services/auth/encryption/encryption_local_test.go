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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLocalKeyEncryptionProvider_WrapUnwrapKey(t *testing.T) {
	const masterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

	if err := os.Setenv("WASMFORGE_AUTH_MASTER_KEY", masterKey); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("WASMFORGE_AUTH_MASTER_KEY") })

	provider, err := NewLocalKeyEncryptionProviderFromEnv("WASMFORGE_AUTH_MASTER_KEY", zap.NewNop())
	require.NoError(t, err)

	dek := []byte("0123456789abcdef0123456789abcdef")
	wrapped, metadata, err := provider.WrapKey(context.Background(), dek)
	require.NoError(t, err)
	assert.NotEmpty(t, wrapped)
	assert.Equal(t, "local", provider.ProviderName())
	assert.NotNil(t, metadata)

	unwrapped, err := provider.UnwrapKey(context.Background(), wrapped, metadata)
	require.NoError(t, err)
	assert.Equal(t, dek, unwrapped)
}

func TestEncryptAndDecryptPrivateKeyPEM_LocalProvider(t *testing.T) {
	const masterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

	if err := os.Setenv("WASMFORGE_AUTH_MASTER_KEY", masterKey); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("WASMFORGE_AUTH_MASTER_KEY") })

	provider, err := NewLocalKeyEncryptionProviderFromEnv("WASMFORGE_AUTH_MASTER_KEY", zap.NewNop())
	require.NoError(t, err)

	payload, err := EncryptPrivateKeyPEM(context.Background(), "-----BEGIN PRIVATE KEY-----\nsecret\n-----END PRIVATE KEY-----", provider)
	require.NoError(t, err)
	assert.NotEmpty(t, payload.Ciphertext)
	assert.NotEmpty(t, payload.WrappedDEK)
	assert.NotEmpty(t, payload.Nonce)
	assert.Equal(t, "AES-256-GCM", payload.Algorithm)
	assert.Equal(t, provider.ProviderName(), payload.Provider)

	plaintext, err := DecryptPrivateKeyPEM(context.Background(), payload, provider)
	require.NoError(t, err)
	assert.Equal(t, "-----BEGIN PRIVATE KEY-----\nsecret\n-----END PRIVATE KEY-----", plaintext)
}
