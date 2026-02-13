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

package route

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	validationutil "github.com/mikhail5545/wasmforge/internal/util/validation"
)

func (req GetRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ID, is.UUID),
		validation.Field(&req.Path, validation.Match(regexp.MustCompile(`^/.*`))),
	)
}

func (req ListRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.IDs, validation.Each(is.UUID)),
		validation.Field(&req.PluginIDs, validation.Each(is.UUID)),
		validation.Field(&req.Paths, validation.Each(validation.Match(regexp.MustCompile(`^/.*`)))),
		validation.Field(&req.TargetURLs, validation.Each(is.URL)),
		validation.Field(&req.OrderField, validation.In(OrderFieldCreatedAt, OrderFieldPath, OrderFieldTargetURL)),
		validation.Field(&req.OrderDirection, validation.In("asc", "desc")),
		validation.Field(&req.PageSize, validation.Required, validation.Min(1), validation.Max(100)),
	)
}

func (req CreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Path, validationutil.PathRule(true)...),
		validation.Field(&req.TargetURL, validation.Required, is.URL),
		validation.Field(&req.IdleConnTimeout, validation.Min(0)),
		validation.Field(&req.TLSHandshakeTimeout, validation.Min(0)),
		validation.Field(&req.ExpectContinueTimeout, validation.Min(0)),
		validation.Field(&req.MaxIdleCons, validation.NilOrNotEmpty, validation.Min(0)),
		validation.Field(&req.MaxIdleConsPerHost, validation.NilOrNotEmpty, validation.Min(0)),
		validation.Field(&req.MaxConsPerHost, validation.NilOrNotEmpty, validation.Min(0)),
		validation.Field(&req.ResponseHeaderTimeout, validation.NilOrNotEmpty, validation.Min(0)),
	)
}

func (req DeleteRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ID, validationutil.UUIDRule(true)...),
	)
}
