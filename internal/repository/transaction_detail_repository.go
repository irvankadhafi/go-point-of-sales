package repository

import (
	"context"
	"github.com/irvankadhafi/go-point-of-sales/cacher"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type transactionDetailRepository struct {
	db    *gorm.DB
	cache cacher.CacheManager
}

func NewTransactionDetailRepository(
	db *gorm.DB,
	cache cacher.CacheManager,
) model.TransactionDetailRepository {
	return &transactionDetailRepository{
		db:    db,
		cache: cache,
	}
}

func (t *transactionDetailRepository) Create(ctx context.Context, tx *gorm.DB, details []*model.TransactionDetail) error {
	if len(details) <= 0 {
		return nil
	}

	if err := tx.WithContext(ctx).Create(details).Error; err != nil {
		logrus.WithFields(logrus.Fields{
			"ctx":     utils.DumpIncomingContext(ctx),
			"details": utils.Dump(details),
		}).Error(err)
		return err
	}

	return nil
}
