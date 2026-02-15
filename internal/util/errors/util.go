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
	"net/http"

	serviceerrors "github.com/mikhail5545/wasmforge/internal/errors"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details any    `json:"details,omitempty"`
	} `json:"error"`
}

// MapServiceError converts service-layer errors into [ErrorResponse] struct for transport layers.
func MapServiceError(err error) (int, ErrorResponse) {
	var resp ErrorResponse

	switch {
	case errors.Is(err, serviceerrors.ErrAlreadyExists):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrAlreadyExists]
		resp.Error.Message = "Already exists"
		resp.Error.Details = err.Error()
		return http.StatusConflict, resp
	case errors.Is(err, serviceerrors.ErrCanceled):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrCanceled]
		resp.Error.Message = "Context canceled"
		resp.Error.Details = err.Error()
		return http.StatusGatewayTimeout, resp
	case errors.Is(err, serviceerrors.ErrConflict):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrConflict]
		resp.Error.Message = "Recourse conflict"
		resp.Error.Details = err.Error()
		return http.StatusConflict, resp
	case errors.Is(err, serviceerrors.ErrInvalidArgument):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrInvalidArgument]
		resp.Error.Message = "Invalid or malformed input"
		resp.Error.Details = err.Error()
		return http.StatusBadRequest, resp
	case errors.Is(err, serviceerrors.ErrNotFound):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrNotFound]
		resp.Error.Message = "Resource not found"
		resp.Error.Details = err.Error()
		return http.StatusNotFound, resp
	case errors.Is(err, serviceerrors.ErrPermissionDenied):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrPermissionDenied]
		resp.Error.Message = "Permission denied"
		resp.Error.Details = err.Error()
		return http.StatusForbidden, resp
	case errors.Is(err, serviceerrors.ErrTooManyRequests):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrTooManyRequests]
		resp.Error.Message = "Rate limit exceeded"
		resp.Error.Details = err.Error()
		return http.StatusTooManyRequests, resp
	case errors.Is(err, serviceerrors.ErrUnimplemented):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrUnimplemented]
		resp.Error.Message = "Not implemented"
		resp.Error.Details = err.Error()
		return http.StatusNotImplemented, resp
	case errors.Is(err, serviceerrors.ErrUnavailable):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrUnavailable]
		resp.Error.Message = "Service unavailable"
		resp.Error.Details = err.Error()
		return http.StatusServiceUnavailable, resp
	case errors.Is(err, serviceerrors.ErrValidationFailed):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrValidationFailed]
		resp.Error.Message = "Unprocessable entity"
		resp.Error.Details = err.Error()
		return http.StatusUnprocessableEntity, resp
	default:
		resp.Error.Code = "INTERNAL_SERVER_ERROR"
		resp.Error.Message = "Internal server error"
		return http.StatusInternalServerError, resp
	}
}
