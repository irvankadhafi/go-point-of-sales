package model

import (
	"gorm.io/gorm"
	"time"
)

// AuditRepository bridges Audit with storage
type AuditRepository interface {
	Audit(*gorm.DB, any, *Audit) (*gorm.DB, error)
}

// Audit represents audit
type Audit struct {
	UserID         int64
	AuditableType  string
	AuditableID    int64
	Action         string
	AuditedChanges string
	CreatedAt      time.Time
}
