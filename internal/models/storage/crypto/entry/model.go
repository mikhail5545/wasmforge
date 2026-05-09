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

package entry

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CryptoMaterialEntryType string

const (
	CryptoMaterialEntryTypeCertificate CryptoMaterialEntryType = "certificate"
	CryptoMaterialEntryTypePublicKey   CryptoMaterialEntryType = "public_key"
	CryptoMaterialEntryTypePrivateKey  CryptoMaterialEntryType = "private_key"
)

type CryptoMaterialEntry struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	ObjectRefBucket      string                  `gorm:"type:varchar(255);not null" json:"object_bucket"`
	ObjectRefKey         string                  `gorm:"type:varchar(512);not null;uniqueIndex" json:"object_key"`
	MaterialID           uuid.UUID               `gorm:"type:uuid;not null" json:"material_id"`
	EntryType            CryptoMaterialEntryType `gorm:"not null" json:"entry_type"`
	Position             int                     `gorm:"not null;default:0;check:position_non_neg,position >= 0" json:"position"`
	FingerprintSHA256Hex string                  `gorm:"type:varchar(128);not null;index" json:"fingerprint_sha256_hex"`
	Algorithm            *string                 `gorm:"type:varchar(128)" json:"algorithm,omitempty"`       // May be null for encrypted private keys
	Details              *string                 `gorm:"type:varchar(256)" json:"details,omitempty"`         // May be null for encrypted private keys
	Subject              *string                 `gorm:"type:varchar(256)" json:"subject,omitempty"`         // Only for certificates
	Issuer               *string                 `gorm:"type:varchar(256)" json:"issuer,omitempty"`          // only for certificates
	SerialHex            *string                 `gorm:"type:varchar(512)" json:"serial_hex,omitempty"`      // Only for certificates
	NotBefore            *time.Time              `gorm:"type:datetime;not null" json:"not_before,omitempty"` // Only for certificates
	NotAfter             *time.Time              `gorm:"type:datetime;not null" json:"not_after,omitempty"`  // Only for certificates
	IsCA                 bool                    `gorm:"not null" json:"is_ca"`                              // Only for certificates
	Checksum             string                  `gorm:"type:varchar(128)" json:"checksum"`
	SizeBytes            int64                   `gorm:"default:0" json:"size_bytes"`

	// Encryption (all null for non-private material)

	WrappedDEK                 *string `gorm:"type:text" json:"-"`
	EncryptionNonce            *string `gorm:"type:text" json:"-"`
	EncryptionAlgorithm        *string `gorm:"type:varchar(64)" json:"-"`
	EncryptionProvider         *string `gorm:"type:varchar(64)" json:"-"`
	EncryptionProviderMetadata *string `gorm:"type:jsonb" json:"-"`

	MetadataJSON map[string]any `gorm:"type:jsonb;default:'{}'" json:"metadata_json,omitempty"`
}

func (*CryptoMaterialEntry) TableName() string {
	return "crypto_material_entries"
}

func (e *CryptoMaterialEntry) BeforeCreate(_ *gorm.DB) (err error) {
	if e.ID == uuid.Nil {
		e.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}
