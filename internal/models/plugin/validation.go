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

package plugin

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	validationutil "github.com/mikhail5545/wasmforge/internal/util/validation"
)

func (req GetRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ID, validationutil.UUIDRule(false)...),
		validation.Field(&req.Filename, validationutil.WasmFilenameRule(false)...),
		validation.Field(&req.Name, validationutil.PluginNameRule(false)...),
	)
}

func (req ListRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.IDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&req.Filenames, validation.Each(validationutil.WasmFilenameRule(false)...)),
		validation.Field(&req.Names, validation.Each(validationutil.PluginNameRule(false)...)),
		validation.Field(&req.OrderField, validation.In(OrderFieldName, OrderFieldFilename, OrderFieldCreatedAt).Error("invalid order field")),
		validation.Field(&req.OrderDirection, validation.In("asc", "desc").Error("invalid order direction")),
		validation.Field(&req.PageSize, validation.Required, validation.Min(1), validation.Max(100)),
	)
}

func (req CreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name, validationutil.PluginNameRule(true)...),
		validation.Field(&req.Filename, validationutil.WasmFilenameRule(true)...),
	)
}

func (req DeleteRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ID, validationutil.UUIDRule(true)...),
	)
}
