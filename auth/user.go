package auth

import (
	"context"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/rbac"
)

type contextKey string

// use module path to make it unique
const userCtxKey contextKey = "github.com/irvankadhafi/go-point-of-sales/auth.User"

// SetUserToCtx set user to context
func SetUserToCtx(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userCtxKey, user)
}

// GetUserFromCtx get user from context
func GetUserFromCtx(ctx context.Context) *User {
	user, ok := ctx.Value(userCtxKey).(User)
	if !ok {
		return nil
	}
	return &user
}

// User represent an authenticated user
type User struct {
	ID             int64                `json:"id"`
	Role           rbac.Role            `json:"role"`
	SessionID      int64                `json:"session_id"`
	RolePermission *rbac.RolePermission `json:"-"`
}

// NewUserFromSession return new user from session
func NewUserFromSession(sess model.Session, perm *rbac.Permission) User {
	rp := rbac.NewRolePermission(sess.Role, perm)
	return User{
		ID:             sess.UserID,
		Role:           sess.Role,
		RolePermission: rp,
		SessionID:      sess.ID,
	}
}

// HasAccess check the user authorization
func (u *User) HasAccess(resource rbac.Resource, action rbac.Action) error {
	if u.RolePermission == nil || !u.RolePermission.HasAccess(resource, action) {
		return ErrAccessDenied
	}

	return nil
}
