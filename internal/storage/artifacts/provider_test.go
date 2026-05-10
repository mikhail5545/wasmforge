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

package artifacts

import (
	"testing"

	"github.com/google/uuid"
	"github.com/mikhail5545/wasmforge/internal/util/memory"
	"github.com/stretchr/testify/assert"
)

func TestLoadOptions_Validate(t *testing.T) {
	id := mustUUID(t)
	validWithID := LoadOptions{
		ID: &id,
	}
	validWithoutID := LoadOptions{
		ProjectID: memory.Ptr(mustUUID(t)),
		Name:      memory.Ptr("test"),
		Version:   memory.Ptr("2.1.0-rc.1"),
	}

	for _, tt := range []struct {
		name      string
		opt       LoadOptions
		expectErr bool
	}{
		{
			name:      "valid with ID",
			opt:       validWithID,
			expectErr: false,
		},
		{
			name:      "valid with project_id, name and version",
			opt:       validWithoutID,
			expectErr: false,
		},
		{
			name: "both",
			opt: LoadOptions{
				ID:        validWithID.ID,
				ProjectID: validWithoutID.ProjectID,
				Name:      validWithoutID.Name,
				Version:   validWithoutID.Version,
			},
			expectErr: false,
		},
		{
			name:      "empty",
			opt:       LoadOptions{},
			expectErr: true,
		},
		{
			name: "invalid",
			opt: LoadOptions{
				ProjectID: memory.Ptr(uuid.Nil),
				Name:      memory.Ptr(""),
				Version:   memory.Ptr(""),
			},
			expectErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opt.Validate()
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func mustUUID(t *testing.T) uuid.UUID {
	t.Helper()

	id, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return id
}
