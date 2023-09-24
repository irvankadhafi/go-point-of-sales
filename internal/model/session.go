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
	ID                    int64
	UserID                int64
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiredAt  time.Time
	RefreshTokenExpiredAt time.Time
	AppID                 int64
	UserAgent             string
	IPAddress             string
	CreatedAt             time.Time
	UpdatedAt             time.Time

	Role rbac.Role `gorm:"-"`
}

// IsAccessTokenExpired check access token expired at against now
func (s *Session) IsAccessTokenExpired() bool {
	return time.Now().After(s.AccessTokenExpiredAt)
}

// NewSessionTokenCacheKey return cache key for session token
func NewSessionTokenCacheKey(token string) string {
	return fmt.Sprintf("cache:id:session_token:%s", token)
}
