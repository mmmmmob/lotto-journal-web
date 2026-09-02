package models

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID                uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OwnerID           *uuid.UUID `gorm:"type:uuid"                                      json:"owner_id,omitempty"`
	StorageKey        string     `gorm:"type:varchar;unique;not null"                   json:"storage_key"`
	FileType          *string    `gorm:"type:varchar"                                   json:"file_type,omitempty"`
	ExtractionStatus  *string    `gorm:"type:varchar(20)"                               json:"extraction_status,omitempty"`
	ExtractedResponse []byte     `gorm:"type:jsonb"                                     json:"extracted_response,omitempty"`
	ExtractionModel   *string    `gorm:"type:varchar(50)"                               json:"extraction_model,omitempty"`
	CreatedAt         time.Time  `gorm:"type:timestamp;autoCreateTime"                  json:"created_at"`
}
