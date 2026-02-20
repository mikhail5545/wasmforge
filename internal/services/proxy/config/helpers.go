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
	configmodel "github.com/mikhail5545/wasmforge/internal/models/proxy/config"
	"github.com/mikhail5545/wasmforge/internal/util/patch"
)

func buildUpdates(existing *configmodel.Config, req *configmodel.UpdateRequest) map[string]any {
	updates := make(map[string]any)

	patch.UpdateIfChanged(updates, "listen_port", req.ListenPort, &existing.ListenPort)
	patch.UpdateIfChanged(updates, "read_header_timeout", req.ReadHeaderTimeout, &existing.ReadHeaderTimeout)
	patch.UpdateIfChanged(updates, "tls_enabled", req.TLSEnabled, &existing.TLSEnabled)
	patch.UpdateIfChanged(updates, "tls_cert_path", req.TLSCertPath, existing.TLSCertPath)
	patch.UpdateIfChanged(updates, "tls_key_path", req.TLSKeyPath, existing.TLSKeyPath)

	return updates
}
