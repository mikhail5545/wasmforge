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

package deployment

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type (
	ExecutionMode  string
	ExecutionPhase string
)

const (
	ExecutionModePlugin ExecutionMode  = "plugin"
	ExecutionPhaseApp   ExecutionPhase = "app"

	ExecutionPhaseOnRequest  ExecutionPhase = "on_request"
	ExecutionPhaseOnResponse ExecutionPhase = "on_response"
	ExecutionPhaseBeforeAuth ExecutionPhase = "before_auth"
	ExecutionPhaseAfterAuth  ExecutionPhase = "after_auth"
	ExecutionPhaseTerminal   ExecutionPhase = "terminal"
)

type Deployment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AppID   uuid.UUID `gorm:"type:uuid;not null;index" json:"app_id"`
	RouteID uuid.UUID `gorm:"type:uuid;not null;index" json:"route_id"`

	AppVersionConstraint string `gorm:"type:varchar(64);not null" json:"app_version_constraint"`

	ExecutionMode  ExecutionMode  `gorm:"type:varchar(32);not null;default:'plugin'" json:"execution_mode"`
	ExecutionPhase ExecutionPhase `gorm:"type:varchar(32);not null;default:'on_request'" json:"execution_phase"`

	Config  datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"config,omitempty"`
	Enabled bool           `gorm:"default:false" json:"enabled"`
}

func (*Deployment) TableName() string {
	return "deployments"
}

func (d *Deployment) BeforeCreate(_ *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}
