package auth

import (
	"github.com/irvankadhafi/go-point-of-sales/rbac"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUser_HasAccess(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		perm := rbac.NewPermission()
		perm.Add(rbac.RoleAdmin, rbac.ResourceUser, rbac.ActionCreateAny)
		user := User{
			ID:             utils.GenerateID(),
			Role:           rbac.RoleAdmin,
			RolePermission: rbac.NewRolePermission(rbac.RoleAdmin, perm),
		}

		err := user.HasAccess(rbac.ResourceUser, rbac.ActionCreateAny)
		require.NoError(t, err)
	})

	t.Run("error access denied", func(t *testing.T) {
		user := User{
			ID:   utils.GenerateID(),
			Role: rbac.RoleCashiers,
		}

		err := user.HasAccess(rbac.ResourceUser, rbac.ActionCreateAny)
		require.Error(t, err)
		require.Equal(t, ErrAccessDenied, err)
	})
}
