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

package core

import (
	"errors"
	"fmt"
)

var (
	ErrSizeLimitExceeded   = errors.New("content size exceeds the configured limit")
	ErrObjectNotFound      = errors.New("object not found")
	ErrInvalidObjectRef    = errors.New("invalid object reference")
	ErrInvalidObjectFormat = errors.New("invalid object format")
	ErrAmbiguousInput      = errors.New("ambiguous input")
)

func NewSizeLimitExceededError(v any) error {
	return fmt.Errorf("%w: %v", ErrSizeLimitExceeded, v)
}

func NewObjectNotFoundError(v any) error {
	return fmt.Errorf("%w: %v", ErrObjectNotFound, v)
}

func NewInvalidObjectRef(v any) error {
	return fmt.Errorf("%w: %v", ErrInvalidObjectRef, v)
}

func NewInvalidObjectFormatError(v any) error {
	return fmt.Errorf("%w: %v", ErrInvalidObjectFormat, v)
}

func NewAmbiguousInputError(v any) error {
	return fmt.Errorf("%w: %v", ErrAmbiguousInput, v)
}
