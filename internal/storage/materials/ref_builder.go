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

package materials

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/mikhail5545/wasmforge/internal/storage/core"
)

type BuildParams struct {
	ProjectID uuid.UUID
	AppID     *uuid.UUID
	ObjectID  uuid.UUID
	Encrypted bool
	Extension string
}

type RefBuilder struct {
	dataRoot string
}

func NewRefBuilder(dataRoot string) *RefBuilder {
	return &RefBuilder{
		dataRoot: dataRoot,
	}
}

func (b *RefBuilder) DataRoot() string {
	return b.dataRoot
}

func (b *RefBuilder) BuildTemp() (core.ObjectRef, error) {
	randID, err := uuid.NewRandom()
	if err != nil {
		return core.ObjectRef{}, fmt.Errorf("error generating random UUID: %v", err)
	}
	return core.ObjectRef{
		Bucket: "tmp",
		Key: filepath.Join(
			b.dataRoot,
			"tmp",
			randID.String()+".tmp",
		),
	}, nil
}

func (b *RefBuilder) Build(params BuildParams) core.ObjectRef {
	if params.Extension == "" {
		params.Extension = ".pem"
	}
	var filename string
	if params.Encrypted {
		filename = fmt.Sprintf("%s.%s.enc", params.ObjectID.String(), params.Extension)
	} else {
		filename = fmt.Sprintf("%s.%s", params.ObjectID.String(), params.Extension)
	}
	if params.AppID != nil {
		return core.ObjectRef{
			Bucket: "certificates",
			Key: filepath.Join(
				b.dataRoot,
				"objects",
				"certificates",
				params.ProjectID.String(),
				params.AppID.String(),
				filename,
			),
		}
	}
	return core.ObjectRef{
		Bucket: "certificates",
		Key: filepath.Join(
			b.dataRoot,
			"objects",
			"certificates",
			params.ProjectID.String(),
			filename,
		),
	}
}
