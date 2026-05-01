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

package audit

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Action string
type Result string

const (
	ActionValidate Action = "validate"
	ActionIssue    Action = "issue"
	ActionRevoke   Action = "revoke"

	ResultSuccess Result = "success"
	ResultFailure Result = "failure"
)

type AuthAudit struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	CreatedAt time.Time `json:"created_at"`

	RouteID           uuid.UUID  `gorm:"type:uuid;not null;index" json:"route_id"`
	AuthConfigID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"auth_config_id"`
	Action            Action     `gorm:"type:varchar(32);not null;index;check:action_checker,action IN ('validate', 'issue', 'revoke')" json:"action"`
	Result            Result     `gorm:"type:varchar(16);not null;check:result_checker,result IN ('success', 'failure')" json:"result"`
	Subject           string     `gorm:"type:varchar(512)" json:"subject,omitempty"`
	TokenJTI          string     `gorm:"type:varchar(256);index" json:"token_jti,omitempty"`
	IssuedAt          *time.Time `json:"issued_at,omitempty"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	ErrorMessage      string     `gorm:"type:text" json:"error_message,omitempty"`
	ClientIP          string     `gorm:"type:varchar(45)" json:"client_ip,omitempty"`
	UserAgent         string     `gorm:"type:text" json:"user_agent,omitempty"`
	AdditionalContext string     `gorm:"type:jsonb" json:"additional_context,omitempty"`
}

func (*AuthAudit) TableName() string {
	return "auth_audits"
}

func (aa *AuthAudit) BeforeCreate(_ *gorm.DB) (err error) {
	if aa.ID == uuid.Nil {
		aa.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}
