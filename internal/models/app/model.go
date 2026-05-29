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

package app

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type (
	Status string
	Type   string
)

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"

	TypeFunction Type = "function"
	TypePlugin   Type = "plugin"
	TypeApp      Type = "app"
)

type App struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	Name        string  `gorm:"type:varchar(128);not null" json:"name"`
	Slug        string  `gorm:"type:varchar(128);not null;uniqueIndex" json:"slug"`
	Description *string `gorm:"type:text" json:"description,omitempty"`

	ProjectID uuid.UUID `gorm:"type:uuid;not null;index" json:"project_id"`

	Status Status `gorm:"type:varchar(32);not null;default:'active'" json:"status"`
	Type   Type   `gorm:"type:varchar(32);not null;default:'function'" json:"type"`

	Metadata datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata,omitempty"`
}

func (*App) TableName() string {
	return "apps"
}

func (a *App) BeforeCreate(_ *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s Status) String() string {
	return string(s)
}

func (t Type) String() string {
	return string(t)
}
