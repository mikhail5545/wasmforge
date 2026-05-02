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

package config

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	validationutil "github.com/mikhail5545/wasmforge/internal/util/validation"
)

func (req *GetRequest) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.RouteID, validationutil.UUIDRule(true)...),
	)
}

func (req *UpsertRequest) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.RouteID, validationutil.UUIDRule(true)...),
		validation.Field(&req.KeyBackendType, validation.Required, validation.In(
			KeyBackendTypeDatabase, KeyBackendTypeJWKS, KeyBackendTypeEnv,
		)),
		validation.Field(&req.TokenTTLSeconds, validation.Required, validation.Min(1)),
		validation.Field(&req.AllowedAlgorithms, validation.Required, validation.Each(validation.In("RS256").Error("only RS256 algorithm is supported"))),
	)
}
