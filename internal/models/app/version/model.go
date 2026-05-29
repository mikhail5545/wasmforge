/*
 * Copyright (c) 2026. Mikhail Kulik
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package version

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AppVersion struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AppID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_app_version,priority:1" json:"app_id"`
	Version string    `gorm:"type:varchar(64);not null;uniqueIndex:idx_app_version,priority:2" json:"version"`

	ArtifactID uuid.UUID `gorm:"type:uuid;not null" json:"artifact_id"`

	Manifest     datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"manifest,omitempty"`
	Capabilities datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"capabilities,omitempty"`
	Limits       datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"limits,omitempty"`
}

func (*AppVersion) TableName() string {
	return "app_versions"
}

func (v *AppVersion) BeforeCreate(_ *gorm.DB) (err error) {
	if v.ID == uuid.Nil {
		v.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}
