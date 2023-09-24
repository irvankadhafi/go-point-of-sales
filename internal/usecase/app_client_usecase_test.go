package usecase

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/internal/model/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAppClientUsecase_FindClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	mockRepo := mock.NewMockAppClientRepository(ctrl)
	ucase := appClientUsecase{appClientRepo: mockRepo}

	respClient := &model.AppClient{
		ID:           100,
		ClientID:     "ic-cms",
		ClientSecret: "12345678",
	}

	t.Run("ok", func(t *testing.T) {
		mockRepo.EXPECT().FindByClientID(ctx, "ic-cms").Times(1).Return(respClient, nil)
		res, err := ucase.FindClient(ctx, "ic-cms", "12345678")
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("handle error", func(t *testing.T) {
		mockRepo.EXPECT().FindByClientID(ctx, "ic-cms").Times(1).Return(nil, errors.New("some repo error"))
		res, err := ucase.FindClient(ctx, "ic-cms", "12345678")
		require.Error(t, err)
		require.NotEqual(t, err, ErrNotFound)
		require.Nil(t, res)
	})
}
