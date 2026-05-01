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

package audit

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	auditmodel "github.com/mikhail5545/wasmforge/internal/models/auth/audit"
	validationutil "github.com/mikhail5545/wasmforge/internal/util/validation"
)

func (f *filter) Validate() error {
	return validation.ValidateStruct(f,
		validation.Field(&f.IDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&f.AuthConfigIDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&f.RouteIDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&f.Actions, validation.Each(validation.In(auditmodel.ActionValidate, auditmodel.ActionRevoke, auditmodel.ActionIssue))),
		validation.Field(&f.Results, validation.Each(validation.In(auditmodel.ResultSuccess, auditmodel.ResultFailure))),
	)
}

func (f *filter) ValidateForList() error {
	if err := f.Validate(); err != nil {
		return err
	}

	return validation.ValidateStruct(f,
		validation.Field(&f.OrderDirection, validation.Required, validation.In("asc", "desc", "ASC", "DESC")),
		validation.Field(&f.OrderField, validation.Required, validation.In(
			auditmodel.OrderFieldCreatedAt, auditmodel.OrderFieldResult,
			auditmodel.OrderFieldAction, auditmodel.OrderFieldClientIP,
			auditmodel.OrderFieldUserAgent,
		)),
		validation.Field(&f.PageSize, validation.Required, validation.Min(5), validation.Max(100)),
	)
}
