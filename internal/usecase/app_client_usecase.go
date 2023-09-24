package usecase

import (
	"context"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
)

type appClientUsecase struct {
	appClientRepo model.AppClientRepository
}

// NewAppClientUsecase :nodoc:
func NewAppClientUsecase(
	appClientRepo model.AppClientRepository,
) model.AppClientUsecase {
	return &appClientUsecase{
		appClientRepo: appClientRepo,
	}
}

// FindClient :nodoc:
func (a *appClientUsecase) FindClient(ctx context.Context, clientID, clientSecret string) (*model.AppClient, error) {
	client, err := a.appClientRepo.FindByClientID(ctx, clientID)
	if err != nil {
		return nil, ErrPermissionDenied
	}
	if client.ClientSecret != clientSecret {
		return nil, ErrPermissionDenied
	}
	return client, nil
}
