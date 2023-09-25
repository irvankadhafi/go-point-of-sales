package usecase

import (
	"context"
	"errors"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/rbac"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/sirupsen/logrus"
	"sync"
)

type transactionUsecase struct {
	transactionRepo model.TransactionRepository
	productRepo     model.ProductRepository
}

// NewTransactionUsecase instantiate a new transaction usecase
func NewTransactionUsecase(transactionRepo model.TransactionRepository, productRepo model.ProductRepository) model.TransactionUsecase {
	return &transactionUsecase{
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
	}
}

func (t *transactionUsecase) Create(ctx context.Context, requester *model.User, input model.CreateTransactionInput) (*model.Transaction, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"requester": utils.Dump(requester),
		"input":     utils.Dump(input),
	})

	if !requester.HasAccess(rbac.ResourceTransaction, rbac.ActionCreateAny) {
		return nil, ErrPermissionDenied
	}

	newTransaction := &model.Transaction{
		ID:         utils.GenerateID(),
		AmountPaid: input.AmountPaid,
		CreatedBy:  requester.ID,
	}

	// Calculate total price and check product stock
	var (
		totalAmount int64
		wg          sync.WaitGroup
	)

	errCh := make(chan error, len(input.TransactionDetails))
	transactionDetails := make([]*model.TransactionDetail, len(input.TransactionDetails))
	for i, detail := range input.TransactionDetails {
		wg.Add(1)

		go func(i int, detail model.TransactionDetail) {
			defer wg.Done()

			product, err := t.productRepo.FindByID(ctx, detail.ProductID)
			if err != nil {
				errCh <- err
				return
			}
			if product == nil {
				errCh <- ErrNotFound
				return
			}

			if product.Quantity < detail.Quantity {
				errCh <- errors.New("insufficient stock for one or more products")
				return
			}

			// Calculate subtotal for transaction detail
			subtotal := product.Price * detail.Quantity
			totalAmount += subtotal

			// Reduce product stock
			product.Quantity -= detail.Quantity
			if err := t.productRepo.Update(ctx, requester.ID, product); err != nil {
				errCh <- err
				return
			}

			detail.TransactionID = newTransaction.ID
			detail.Subtotal = subtotal
			transactionDetails[i] = &detail
		}(i, detail)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		return nil, err
	}

	newTransaction.TotalPrice = totalAmount
	newTransaction.Change = newTransaction.AmountPaid - newTransaction.TotalPrice
	newTransaction.TransactionDetails = transactionDetails

	// Save the transaction to the repository
	if err := t.transactionRepo.Create(ctx, requester.ID, newTransaction); err != nil {
		logger.Error(err)
		return nil, err
	}

	return newTransaction, nil
}
