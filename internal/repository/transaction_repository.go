package repository

import (
	"context"
	"github.com/irvankadhafi/go-point-of-sales/cacher"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

type transactionRepository struct {
	db                    *gorm.DB
	cache                 cacher.CacheManager
	transactionDetailRepo model.TransactionDetailRepository
	auditRepo             model.AuditRepository
}

func NewTransactionRepository(
	db *gorm.DB,
	cache cacher.CacheManager,
	transactionDetailRepo model.TransactionDetailRepository,
	auditRepo model.AuditRepository,
) model.TransactionRepository {
	return &transactionRepository{
		db:                    db,
		cache:                 cache,
		transactionDetailRepo: transactionDetailRepo,
		auditRepo:             auditRepo,
	}
}

func (t *transactionRepository) Create(ctx context.Context, userID int64, transaction *model.Transaction) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":         utils.DumpIncomingContext(ctx),
		"userID":      userID,
		"transaction": utils.Dump(transaction),
	})

	err := t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(transaction).Error; err != nil {
			logger.Error(err)
			return err
		}

		if err := t.transactionDetailRepo.Create(ctx, tx, transaction.TransactionDetails); err != nil {
			logger.Error(err)
			return err
		}

		if err := t.auditRepo.Audit(ctx, tx, transaction, &model.Audit{
			UserID:        userID,
			AuditableType: t.name(),
			AuditableID:   transaction.ID,
			Action:        model.AuditActionCreate,
			CreatedAt:     time.Now(),
		}); err != nil {
			logger.Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (t *transactionRepository) name() string {
	return "transaction"
}
