package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Threat is a confirmed malicious finding, regardless of source.
type Threat struct {
	ID         string    `gorm:"primaryKey;type:text"`
	Name       string    `gorm:"type:text;not null"`
	Type       string    `gorm:"type:text"` // virus | malware | ransomware | pua | rootkit
	Severity   string    `gorm:"type:text"` // low | medium | high | critical
	FilePath   string    `gorm:"type:text"`
	Source     string    `gorm:"type:text"` // file | network | process
	Status     string    `gorm:"type:text"` // active | quarantined | resolved
	DetectedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// QuarantineEntry tracks a file that has been moved to the quarantine vault.
type QuarantineEntry struct {
	ID             string     `gorm:"primaryKey;type:text"`
	ThreatID       string     `gorm:"type:text;index"`
	OriginalPath   string     `gorm:"type:text;not null"`
	QuarantinePath string     `gorm:"type:text;not null"`
	HashSHA256     string     `gorm:"type:text"`
	FileSize       int64      `gorm:"default:0"`
	QuarantinedAt  time.Time
	RestoredAt     *time.Time // nil until restored
}

func (t *Threat) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	return nil
}

func (q *QuarantineEntry) BeforeCreate(tx *gorm.DB) error {
	if q.ID == "" {
		q.ID = uuid.NewString()
	}
	return nil
}
