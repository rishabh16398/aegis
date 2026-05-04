package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ScanJob tracks one scan operation (a file, directory, or full scan).
// One job → many ScanResults.
type ScanJob struct {
	ID           string     `gorm:"primaryKey;type:text"`
	Type         string     `gorm:"type:text;not null"` // file | directory | quick | full
	Status       string     `gorm:"type:text;not null"` // pending | running | completed | cancelled | failed
	Path         string     `gorm:"type:text"`
	TotalFiles   int        `gorm:"default:0"`
	ScannedFiles int        `gorm:"default:0"`
	ThreatsFound int        `gorm:"default:0"`
	StartedAt    *time.Time
	CompletedAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Results []ScanResult `gorm:"foreignKey:JobID"`
}

// ScanResult is one file's outcome within a ScanJob.
type ScanResult struct {
	ID         string    `gorm:"primaryKey;type:text"`
	JobID      string    `gorm:"type:text;not null;index"`
	FilePath   string    `gorm:"type:text;not null"`
	Status     string    `gorm:"type:text;not null"` // clean | threat | error
	ThreatName string    `gorm:"type:text"`
	ThreatType string    `gorm:"type:text"` // virus | malware | ransomware | pua
	Engine     string    `gorm:"type:text"` // clamav | yara
	HashMD5    string    `gorm:"type:text"`
	HashSHA256 string    `gorm:"type:text"`
	FileSize   int64     `gorm:"default:0"`
	ScannedAt  time.Time
}

func (s *ScanJob) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	return nil
}

func (s *ScanResult) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	return nil
}
