package model

import (
	"context"
	"github.com/irvankadhafi/go-point-of-sales/rbac"
)

// LoginRequest request
type LoginRequest struct {
	AppID                                      int64
	Email, PlainPassword, IPAddress, UserAgent string
}

// RefreshTokenRequest request
type RefreshTokenRequest struct {
	AppID                              int64
	RefreshToken, IPAddress, UserAgent string
}

// AuthUsecase usecases
type AuthUsecase interface {
	LoginByEmailPassword(ctx context.Context, req LoginRequest) (*Session, error)
	// AuthenticateToken authenticate the given token
	AuthenticateToken(ctx context.Context, accessToken string) (*User, error)
	FindRolePermission(ctx context.Context, role rbac.Role) (*rbac.RolePermission, error)
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*Session, error)
	DeleteSessionByID(ctx context.Context, sessionID int64) error
}
