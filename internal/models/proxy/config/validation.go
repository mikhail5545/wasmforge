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

import validation "github.com/go-ozzo/ozzo-validation/v4"

func (req UpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ListenPort, validation.NilOrNotEmpty, validation.Min(1), validation.Max(65535)),
		validation.Field(&req.ReadHeaderTimeout, validation.NilOrNotEmpty, validation.Min(1)),

		validation.Field(&req.TLSEnabled, validation.NilOrNotEmpty),
		validation.Field(&req.TLSCertPath, validation.NilOrNotEmpty, validation.When(req.TLSEnabled != nil && *req.TLSEnabled, validation.Required)),
		validation.Field(&req.TLSKeyPath, validation.NilOrNotEmpty, validation.When(req.TLSEnabled != nil && *req.TLSEnabled, validation.Required)),
	)
}
