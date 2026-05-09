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
	"fmt"
)

type KeyProvider interface {
	WrapKey(ctx context.Context, dek []byte) ([]byte, map[string]any, error)
	UnwrapKey(ctx context.Context, wrapped []byte, metadata map[string]any) ([]byte, error)
	Name() string
}

type KeyEncryptionRegistry struct {
	providers map[string]KeyProvider
}

func NewKeyEncryptionRegistry(providers ...KeyProvider) *KeyEncryptionRegistry {
	registry := &KeyEncryptionRegistry{
		providers: make(map[string]KeyProvider, len(providers)),
	}
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		registry.providers[provider.Name()] = provider
	}
	return registry
}

func (r *KeyEncryptionRegistry) Resolve(name string) (KeyProvider, error) {
	if r == nil {
		return nil, fmt.Errorf("key encryption registry is not configured")
	}
	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("key encryption provider %q is not configured", name)
	}
	return provider, nil
}
