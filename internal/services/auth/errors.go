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

package auth

import (
	"errors"
	"fmt"
)

type KeyError struct {
	Code    string
	Message string
	KeyID   string
	Err     error
}

func (e *KeyError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %s: %v", e.Code, e.Message, e.KeyID, e.Err)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.KeyID)
}

func (e *KeyError) Unwrap() error {
	return e.Err
}

const (
	ErrCodeKeyNotFound    = "KEY_NOT_FOUND"
	ErrCodeKeyExpired     = "KEY_EXPIRED"
	ErrCodeInvalidFormat  = "INVALID_KEY_FORMAT"
	ErrCodeInvalidBackend = "INVALID_BACKEND"
	ErrCodeFetchFailed    = "FETCH_FAILED"
)

func NewKeyNotFoundError(keyID string) *KeyError {
	return &KeyError{
		Code:    ErrCodeKeyNotFound,
		Message: "key not found",
		KeyID:   keyID,
	}
}

func NewKeyExpiredError(keyID string) *KeyError {
	return &KeyError{
		Code:    ErrCodeKeyExpired,
		Message: "key has expired",
		KeyID:   keyID,
	}
}

func NewInvalidKeyFormatError(keyID string, err error) *KeyError {
	return &KeyError{
		Code:    ErrCodeInvalidFormat,
		Message: "invalid key format",
		KeyID:   keyID,
		Err:     err,
	}
}

func NewInvalidBackendError(backend string) *KeyError {
	return &KeyError{
		Code:    ErrCodeInvalidBackend,
		Message: "invalid backend type",
		KeyID:   backend,
	}
}

func NewFetchFailedError(keyID string, err error) *KeyError {
	return &KeyError{
		Code:    ErrCodeFetchFailed,
		Message: "failed to fetch key",
		KeyID:   keyID,
		Err:     err,
	}
}

// Token validation and issuance errors
var (
	// Validator errors
	ErrTokenExpired        = errors.New("token expired")
	ErrTokenInvalid        = errors.New("token invalid")
	ErrTokenMalformed      = errors.New("token malformed")
	ErrInvalidSignature    = errors.New("invalid signature")
	ErrMissingClaims       = errors.New("missing required claims")
	ErrInvalidAudience     = errors.New("invalid audience")
	ErrInvalidIssuer       = errors.New("invalid issuer")
	ErrAlgorithmNotAllowed = errors.New("algorithm not allowed")

	// Issuer errors
	ErrIssuerKeyNotFound   = errors.New("issuer key not found")
	ErrInvalidClaimsFormat = errors.New("invalid claims format")
)

func NewTokenExpiredError(msg string) error {
	return fmt.Errorf("%w: %s", ErrTokenExpired, msg)
}

func NewTokenInvalidError(msg string) error {
	return fmt.Errorf("%w: %s", ErrTokenInvalid, msg)
}

func NewTokenMalformedError(msg string) error {
	return fmt.Errorf("%w: %s", ErrTokenMalformed, msg)
}

func NewInvalidSignatureError(msg string) error {
	return fmt.Errorf("%w: %s", ErrInvalidSignature, msg)
}

func NewMissingClaimsError(claims string) error {
	return fmt.Errorf("%w: %s", ErrMissingClaims, claims)
}

func NewInvalidAudienceError(audience string) error {
	return fmt.Errorf("%w: %s", ErrInvalidAudience, audience)
}

func NewInvalidIssuerError(issuer string) error {
	return fmt.Errorf("%w: %s", ErrInvalidIssuer, issuer)
}

func NewAlgorithmNotAllowedError(algorithm string) error {
	return fmt.Errorf("%w: %s", ErrAlgorithmNotAllowed, algorithm)
}

func NewIssuerKeyNotFoundError(msg string) error {
	return fmt.Errorf("%w: %s", ErrIssuerKeyNotFound, msg)
}

func NewInvalidClaimsFormatError(msg string) error {
	return fmt.Errorf("%w: %s", ErrInvalidClaimsFormat, msg)
}
