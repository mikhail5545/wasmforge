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

package stats

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	validationutil "github.com/mikhail5545/wasmforge/internal/util/validation"
)

func (req OverviewRequest) Validate() error {
	return validation.ValidateStruct(&req)
}

func (req RoutesRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Limit, validation.NilOrNotEmpty, validation.Min(1), validation.Max(200)),
	)
}

func (req RouteSummaryRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Path, validationutil.PathRule(true)...),
	)
}

func (req RoutePluginsRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Path, validationutil.PathRule(true)...),
	)
}

func (req TimeseriesRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.BucketSeconds, validation.NilOrNotEmpty, validation.Min(1), validation.Max(3600)),
	)
}
