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
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	materialmodel "github.com/mikhail5545/wasmforge/internal/models/auth/key"
	validationutil "github.com/mikhail5545/wasmforge/internal/util/validation"
)

func (f *filter) Validate() error {
	return validation.ValidateStruct(f,
		validation.Field(&f.IDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&f.AuthConfigIDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&f.RouteIDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&f.Types, validation.Each(validation.In(materialmodel.TypePrivate, materialmodel.TypePublic))),
		validation.Field(&f.ExternalKeyURLs, validation.Each(is.URL)),
	)
}

func (f *filter) ValidateForList() error {
	if err := f.Validate(); err != nil {
		return err
	}

	return validation.ValidateStruct(f,
		validation.Field(&f.OrderField, validation.Required, validation.In(
			materialmodel.OrderFieldAlgorithm, materialmodel.OrderFieldIsActive,
			materialmodel.OrderFieldCreatedAt, materialmodel.OrderFieldType,
		)),
		validation.Field(&f.OrderDirection, validation.Required, validation.In("asc", "desc", "ASC", "DESC")),
		validation.Field(&f.PageSize, validation.Required, validation.Min(5), validation.Max(100)),
	)
}
