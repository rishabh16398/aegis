package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NetworkEvent is a single suspicious or blocked network packet/connection.
type NetworkEvent struct {
	ID          string    `gorm:"primaryKey;type:text"`
	SrcIP       string    `gorm:"type:text"`
	SrcPort     int
	DstIP       string    `gorm:"type:text"`
	DstPort     int
	Protocol    string    `gorm:"type:text"` // tcp | udp | icmp
	Direction   string    `gorm:"type:text"` // inbound | outbound
	ThreatType  string    `gorm:"type:text"`
	RuleMatched string    `gorm:"type:text"` // which Suricata rule fired
	Blocked     bool      `gorm:"default:false"`
	OccurredAt  time.Time
}

// BlockedIP is a manually or automatically blocked IP address.
type BlockedIP struct {
	ID        string     `gorm:"primaryKey;type:text"`
	IP        string     `gorm:"type:text;uniqueIndex;not null"`
	Reason    string     `gorm:"type:text"`
	BlockedAt time.Time
	ExpiresAt *time.Time // nil = permanent
}

func (n *NetworkEvent) BeforeCreate(tx *gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.NewString()
	}
	return nil
}

func (b *BlockedIP) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	return nil
}
