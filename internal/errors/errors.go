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

package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidArgument invalid argument passed, invalid argument format error.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrValidationFailed argument is well-formatted, but conflicts with the validation rules error.
	ErrValidationFailed = errors.New("validation failed")
	// ErrNotFound resource not found error.
	ErrNotFound = errors.New("not found")
	// ErrConflict resource state conflict error.
	ErrConflict = errors.New("state conflict")
	// ErrAlreadyExists resource already exists error.
	ErrAlreadyExists = errors.New("already exists")
	// ErrPermissionDenied caller is not allowed to use this error.
	ErrPermissionDenied = errors.New("permission denied")
	// ErrTooManyRequests request is rate limited error.
	ErrTooManyRequests = errors.New("too many requests")
	// ErrUnimplemented functionality is not implemented error.
	ErrUnimplemented = errors.New("unimplemented")
	// ErrCanceled request context cancelled error.
	ErrCanceled = errors.New("context cancelled")
	// ErrUnavailable external service error.
	ErrUnavailable = errors.New("unavailable")
)

var ErrorAliases = map[error]string{
	ErrInvalidArgument:  "INVALID_ARGUMENT",
	ErrValidationFailed: "VALIDATION_FAILED",
	ErrNotFound:         "NOT_FOUND",
	ErrConflict:         "CONFLICT",
	ErrAlreadyExists:    "ALREADY_EXISTS",
	ErrPermissionDenied: "PERMISSION_DENIED",
	ErrTooManyRequests:  "TOO_MANY_REQUESTS",
	ErrUnimplemented:    "UNIMPLEMENTED",
	ErrCanceled:         "CANCELED",
	ErrUnavailable:      "UNAVAILABLE",
}

func NewInvalidArgumentError(v any) error {
	return fmt.Errorf("%w: %v", ErrInvalidArgument, v)
}

func NewNotFoundError(v any) error {
	return fmt.Errorf("%w: %v", ErrNotFound, v)
}

func NewConflictError(v any) error {
	return fmt.Errorf("%w: %v", ErrConflict, v)
}

func NewAlreadyExistsError(v any) error {
	return fmt.Errorf("%w: %v", ErrAlreadyExists, v)
}

func NewPermissionDeniedError(v any) error {
	return fmt.Errorf("%w: %v", ErrPermissionDenied, v)
}

func NewTooManyRequestsError(v any) error {
	return fmt.Errorf("%w: %v", ErrTooManyRequests, v)
}

func NewUnimplementedError(v any) error {
	return fmt.Errorf("%w: %v", ErrUnimplemented, v)
}

func NewCanceledError(v any) error {
	return fmt.Errorf("%w: %v", ErrCanceled, v)
}

func NewUnavailableError(v any) error {
	return fmt.Errorf("%w: %v", ErrUnavailable, v)
}

func NewValidationError(v any) error {
	return fmt.Errorf("%w: %v", ErrValidationFailed, v)
}
