package delivery

import (
	"context"
	"github.com/irvankadhafi/go-point-of-sales/auth"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
)

// GetAuthUserFromCtx ..
func GetAuthUserFromCtx(ctx context.Context) *model.User {
	authUser := auth.GetUserFromCtx(ctx)
	if authUser == nil {
		return nil
	}

	user := &model.User{
		ID:        authUser.ID,
		Role:      authUser.Role,
		SessionID: authUser.SessionID,
	}

	user.SetRolePermission(authUser.RolePermission)
	return user
}
