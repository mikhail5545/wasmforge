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

package artifact

import (
	"time"

	"github.com/google/uuid"
)

type Role string
type Status string

const (
	RolePlugin Role = "plugin"
	RoleApp    Role = "app"

	StatusUploaded   Status = "uploaded"
	StatusValidated  Status = "validated"
	StatusActive     Status = "active"
	StatusDeprecated Status = "deprecated"
	StatusFailed     Status = "failed"
)

type Artifact struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AppID     *uuid.UUID `gorm:"type:uuid" json:"app_id,omitempty"`
	ProjectID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_artifact_project_name_version_role,priority:1" json:"project_id"`

	Name    string `gorm:"type:varchar(128);not null;uniqueIndex:idx_artifact_project_name_version_role,priority:2" json:"name"`
	Version string `gorm:"type:varchar(64);not null;default:0.0.0;uniqueIndex:idx_artifact_project_name_version_role,priority:3" json:"version"`
	Role    Role   `gorm:"type:varchar(32);not null;uniqueIndex:idx_artifact_project_name_version_role,priority:4" json:"role"`

	Status Status `gorm:"type:varchar(32);not null;default:'uploaded'" json:"status"`

	ContentType string  `gorm:"type:varchar(64);default:'application/wasm'" json:"content_type"`
	Entrypoint  *string `gorm:"type:varchar(128)" json:"entrypoint,omitempty"` // Later this will be centralized from runtime core - 'on_request', 'on_response', 'on_event', etc.

	ObjectRefBucket string `gorm:"type:varchar(255);not null;uniqueIndex:id_artifact_object_ref,priority:2" json:"object_bucket"`
	ObjectRefKey    string `gorm:"type:varchar(1024);not null;uniqueIndex:id_artifact_object_ref,priority:1" json:"object_key"`

	ChecksumSHA256Hex string `gorm:"type:varchar(128);not null;index" json:"checksum_sha_256_hex"`
	SizeBytes         int64  `gorm:"default:0" json:"size_bytes"`

	Metadata string `gorm:"type:jsonb" json:"metadata,omitempty"`
}
