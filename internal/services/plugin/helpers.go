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
	"github.com/google/uuid"
	pluginrepo "github.com/mikhail5545/wasmforge/internal/database/plugin"
	serviceerrors "github.com/mikhail5545/wasmforge/internal/errors"
	pluginmodel "github.com/mikhail5545/wasmforge/internal/models/plugin"
	"github.com/mikhail5545/wasmforge/internal/uploads"
)

func extractIdentifier(req *pluginmodel.GetRequest) (pluginrepo.FilterOption, error) {
	if err := req.Validate(); err != nil {
		return nil, serviceerrors.NewValidationError(err)
	}
	switch {
	case req.ID != nil:
		return pluginrepo.WithIDs(uuid.MustParse(*req.ID)), nil
	case req.Name != nil:
		return pluginrepo.WithNames(*req.Name), nil
	case req.Filename != nil:
		return pluginrepo.WithFilenames(*req.Filename), nil
	default:
		return nil, serviceerrors.NewValidationError("at least one of id, name or filename must be provided")
	}
}

func (s *Service) deleteFile(filename *string) error {
	if filename == nil {
		return nil
	}
	return s.uploadManager.Delete(*filename, uploads.PluginUpload)
}
