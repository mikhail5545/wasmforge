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

package key

import (
	"crypto/x509"
	"encoding/pem"

	inerrors "github.com/mikhail5545/wasmforge/internal/errors"
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	"github.com/mikhail5545/wasmforge/internal/services/auth/metadata"
)

func toKeyResponses(keys []*materialmodel.Material, includePrivate bool) []*materialmodel.Response {
	responses := make([]*materialmodel.Response, 0, len(keys))
	for _, key := range keys {
		responses = append(responses, toKeyResponse(key, includePrivate))
	}
	return responses
}

func toKeyResponse(key *materialmodel.Material, includePrivate bool) *materialmodel.Response {
	var meta map[string]any
	if strings, err := metadata.ParseKeyMetadata(key); err == nil && strings.Primary {
		meta = map[string]any{metadata.Primary: true}
	}

	response := &materialmodel.Response{
		ID:             key.ID.String(),
		KeyID:          key.KeyID,
		CreatedAt:      key.CreatedAt,
		ExpiresAt:      key.ExpiresAt,
		PublicKeyPEM:   key.PublicKeyPEM,
		IsActive:       key.IsActive,
		Algorithm:      key.Algorithm,
		Type:           key.Type,
		AuthConfigID:   key.AuthConfigID.String(),
		ExternalKeyKID: key.ExternalKeyKID,
		ExternalKeyURL: key.ExternalKeyURL,
		Metadata:       meta,
	}
	if includePrivate {
		response.PrivateKeyPEM = key.PrivateKeyPEM
	}
	return response
}

func validatePEMKeys(privateKeyPEM, publicKeyPEM string) error {
	privBlock, _ := pem.Decode([]byte(privateKeyPEM))
	if privBlock == nil {
		return inerrors.NewInvalidArgumentError("invalid private_key_pem format")
	}
	if _, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes); err != nil {
		if _, err = x509.ParsePKCS8PrivateKey(privBlock.Bytes); err != nil {
			return inerrors.NewInvalidArgumentError("failed to parse private key: " + err.Error())
		}
	}

	pubBlock, _ := pem.Decode([]byte(publicKeyPEM))
	if pubBlock == nil {
		return inerrors.NewInvalidArgumentError("invalid public_key_pem format")
	}
	if _, err := x509.ParsePKIXPublicKey(pubBlock.Bytes); err != nil {
		return inerrors.NewInvalidArgumentError("failed to parse public key: " + err.Error())
	}
	return nil
}
