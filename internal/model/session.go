package model

import (
	"context"
	"fmt"
	"github.com/irvankadhafi/go-point-of-sales/rbac"
	"time"
)

// SessionRepository repository for Session
type SessionRepository interface {
	Create(ctx context.Context, sess *Session) error
	FindByToken(ctx context.Context, tokenType TokenType, token string) (*Session, error)
	FindByID(ctx context.Context, id int64) (*Session, error)
	CheckToken(ctx context.Context, token string) (exist bool, err error)
	RefreshToken(ctx context.Context, oldSess, sess *Session) (*Session, error)
	DeleteByUserIDAndMaxRemainderSession(ctx context.Context, userID int64, maxRemainderSess int) error
	Delete(ctx context.Context, session *Session) error
}

// TokenType type of token
type TokenType int

// TokenType constants
const (
	AccessToken  TokenType = 0
	RefreshToken TokenType = 1
)

// Session the user's session
type Session struct {
	ID                    int64     `json:"id" gorm:"primary_key"`
	UserID                int64     `json:"user_id"`
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiredAt  time.Time `json:"access_token_expired_at"`
	RefreshTokenExpiredAt time.Time `json:"refresh_token_expired_at"`
	UserAgent             string    `json:"user_agent"`
	AppID                 int64     `json:"app_id"`
	IPAddress             string    `json:"ip_address"`
	CreatedAt             time.Time `json:"created_at" sql:"DEFAULT:'now()':::STRING::TIMESTAMP" gorm:"->;<-:create"`
	UpdatedAt             time.Time `json:"updated_at" sql:"DEFAULT:'now()':::STRING::TIMESTAMP"`
	Role                  rbac.Role `json:"role" gorm:"-"`
}

// IsAccessTokenExpired check access token expired at against now
func (s *Session) IsAccessTokenExpired() bool {
	return time.Now().After(s.AccessTokenExpiredAt)
}

// NewSessionTokenCacheKey return cache key for session token
func NewSessionTokenCacheKey(token string) string {
	return fmt.Sprintf("cache:id:session_token:%s", token)
}
