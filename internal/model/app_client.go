package model

import (
	"context"
	"time"
)

// AppClient the app clients
type AppClient struct {
	ID           int64     `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AppClientRepository repository
type AppClientRepository interface {
	FindByClientID(ctx context.Context, clientID string) (*AppClient, error)
	FindByID(ctx context.Context, appID int64) (*AppClient, error)
	Create(ctx context.Context, appClient *AppClient) error
}

// AppClientUsecase usecase
type AppClientUsecase interface {
	FindClient(ctx context.Context, clientID, clientSecret string) (*AppClient, error)
}
