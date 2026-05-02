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
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	validationutil "github.com/mikhail5545/wasmforge/internal/util/validation"
)

func (req *GetRequest) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.KeyID, validation.Required),
	)
}

func (req *DeleteRequest) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.KeyID, validation.Required),
	)
}

func (req *ListRequest) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.IDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&req.RouteIDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&req.AuthConfigIDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&req.Types, validation.Each(validation.In(TypePrivate, TypePublic))),
		validation.Field(&req.OrderField, validation.Required, validation.In(
			OrderFieldIsActive, OrderFieldAlgorithm,
			OrderFieldType, OrderFieldCreatedAt,
		)),
		validation.Field(&req.OrderDirection, validation.Required, validation.In("asc", "desc", "ASC", "DESC")),
		validation.Field(&req.PageSize, validation.Required, validation.Min(5), validation.Max(100)),
	)
}

func (req *CreateRequest) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.RouteID, validationutil.UUIDRule(true)...),
		validation.Field(&req.KeyID, validation.Required),
		validation.Field(&req.PublicKeyPEM, validation.Required),
		validation.Field(&req.PrivateKeyPEM, validation.Required),
		validation.Field(&req.ExpiresAt, validation.When(req.ExpiresAt != nil, validation.By(validateExpiresAt))),
		validation.Field(&req.Metadata, is.JSON),
	)
}

func (req *GenerateRequest) Validate() error {
	return validation.ValidateStruct(req,
		validation.Field(&req.KeyID, validation.Required),
		validation.Field(&req.RouteID, validationutil.UUIDRule(true)...),
		validation.Field(&req.ExpiresInDays, validation.Required, validation.Min(1)),
		validation.Field(&req.Metadata, is.JSON),
	)
}

func validateExpiresAt(value any) error {
	return validationutil.IsValidTimeAfterNow(value, time.Duration(2)*time.Hour)
}
