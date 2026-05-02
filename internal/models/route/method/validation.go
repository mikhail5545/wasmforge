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

package method

import (
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	validationutil "github.com/mikhail5545/wasmforge/internal/util/validation"
)

func (req *GetRequest) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.RouteID, validationutil.UUIDRule(true)...),
		validation.Field(&req.Method, validation.Required, validation.In(
			http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch,
			http.MethodHead, http.MethodOptions, http.MethodConnect, http.MethodTrace,
		)),
	)
}

func (req *ListRequest) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.RouteID, validationutil.UUIDRule(true)...),
	)
}

func (req *SetRequest) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.RouteID, validationutil.UUIDRule(true)...),
		validation.Field(&req.Methods, validation.Each(validation.By(validateSetRequestMethodSpec))),
	)
}

func validateSetRequestMethodSpec(value any) error {
	spec, ok := value.(SetRequestMethodSpec)
	if !ok {
		return validation.NewError("validation_invalid_method_spec", "invalid method spec")
	}
	return spec.Validate()
}

func (req *DeleteRequest) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.RouteID, validationutil.UUIDRule(true)...),
		validation.Field(&req.Method, validation.Required, validation.In(
			http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch,
			http.MethodHead, http.MethodOptions, http.MethodConnect, http.MethodTrace,
		)),
	)
}
