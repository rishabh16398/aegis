package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProcessEvent is logged when a running process looks suspicious.
type ProcessEvent struct {
	ID         string    `gorm:"primaryKey;type:text"`
	PID        int       `gorm:"not null"`
	Name       string    `gorm:"type:text"`
	Path       string    `gorm:"type:text"`
	Reason     string    `gorm:"type:text"` // high_cpu | high_mem | suspicious_name | injection
	CPUPercent float64
	MemMB      float64
	DetectedAt time.Time
}

func (p *ProcessEvent) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	return nil
}
