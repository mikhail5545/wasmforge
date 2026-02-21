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
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func (req UpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ListenPort, validation.NilOrNotEmpty, validation.Min(1), validation.Max(65535)),
		validation.Field(&req.ReadHeaderTimeout, validation.NilOrNotEmpty, validation.Min(1)),
	)
}

func (req GenerateCertificatesRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.CommonName, validation.Required, validation.Match(regexp.MustCompile(`^[a-z0-9]+(?:_[&?a-z0-9]+)*$`))),
		validation.Field(&req.ValidDays, validation.Required, validation.Min(1)),
		validation.Field(&req.RsaBits, validation.Required, validation.In(2048, 4096)),
	)
}
