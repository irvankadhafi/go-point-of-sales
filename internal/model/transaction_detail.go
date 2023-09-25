package model

import (
	"context"
	"gorm.io/gorm"
)

type TransactionDetail struct {
	TransactionID int64 `json:"transaction_id"`
	ProductID     int64 `json:"product_id"`
	Quantity      int64 `json:"quantity"`
	Subtotal      int64 `json:"subtotal" sql:"type:decimal(20,0)" gorm:"type:numeric(20,0)"`
}

type TransactionDetailRepository interface {
	Create(ctx context.Context, tx *gorm.DB, detail []*TransactionDetail) error
}
