package model

import (
	"context"
	"gorm.io/gorm"
	"time"
)

// AuditRepository bridges Audit with storage
type AuditRepository interface {
	Audit(context.Context, *gorm.DB, any, *Audit) error
}

// Audit represents audit
type Audit struct {
	UserID         int64       `json:"user_id"`
	AuditableType  string      `json:"auditable_type"`
	AuditableID    int64       `json:"auditable_id"`
	Action         AuditAction `json:"action"`
	AuditedChanges string      `json:"audited_changes"`
	CreatedAt      time.Time   `json:"created_at" sql:"DEFAULT:'now()':::STRING::TIMESTAMP" gorm:"->;<-:create"`
}

type AuditAction string

// AuditAction constants
const (
	AuditActionCreate  AuditAction = "create"
	AuditActionUpdate  AuditAction = "update"
	AuditActionDelete  AuditAction = "delete"
	AuditActionReset   AuditAction = "reset"
	AuditActionUpsert  AuditAction = "upsert"
	AuditActionRestore AuditAction = "restore"
)
