package repository

import (
	"context"
	"encoding/json"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type auditRepo struct{}

// NewAuditRepository create new repository
func NewAuditRepository() model.AuditRepository {
	return &auditRepo{}
}

// Audit :nodoc:
func (r *auditRepo) Audit(ctx context.Context, tx *gorm.DB, item any, audit *model.Audit) error {
	changes, err := json.Marshal(item)
	if err != nil {
		log.WithFields(log.Fields{
			"changes": utils.Dump(item)}).
			Error(err)
		return err
	}

	audit.AuditedChanges = string(changes)

	if err := tx.WithContext(ctx).Create(audit).Error; err != nil {
		log.WithField("audit", utils.Dump(audit)).Error(err)
		return err
	}

	return nil
}
