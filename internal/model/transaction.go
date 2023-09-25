package model

import (
	"context"
	"time"
)

type Transaction struct {
	ID         int64     `json:"id"`
	TotalPrice int64     `json:"total_price" sql:"type:decimal(20,0)" gorm:"type:numeric(20,0)"`
	AmountPaid int64     `json:"amount_paid" sql:"type:decimal(20,0)" gorm:"type:numeric(20,0)"`
	Change     int64     `json:"change" sql:"type:decimal(20,0)" gorm:"type:numeric(20,0)"`
	CreatedBy  int64     `json:"created_by" gorm:"->;<-:create"`                                           // create & read only
	CreatedAt  time.Time `json:"created_at" sql:"DEFAULT:'now()':::STRING::TIMESTAMP" gorm:"->;<-:create"` // create & read only

	TransactionDetails []*TransactionDetail `json:"transaction_details" gorm:"-"`
}

type TransactionRepository interface {
	Create(ctx context.Context, userID int64, transaction *Transaction) error
}

type TransactionUsecase interface {
	Create(ctx context.Context, requester *User, input CreateTransactionInput) (*Transaction, error)
}

// CreateTransactionInput to create a new transaction
type CreateTransactionInput struct {
	TransactionDetails []TransactionDetail `json:"transaction_details"`
	AmountPaid         int64               `json:"amount_paid"`
}
