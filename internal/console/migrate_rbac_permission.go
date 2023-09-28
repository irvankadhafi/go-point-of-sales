package console

import (
	"github.com/irvankadhafi/go-point-of-sales/internal/db"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/internal/repository"
	"github.com/irvankadhafi/go-point-of-sales/rbac"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var migrateRBACPermissionCmd = &cobra.Command{
	Use:   "migrate-rbac-permission",
	Short: "migrate rbac permission",
	Long:  `This subcommand used to migrate database`,
	Run:   migrateRBACPermission,
}

func init() {
	RootCmd.AddCommand(migrateRBACPermissionCmd)
}

func migrateRBACPermission(cmd *cobra.Command, args []string) {
	db.InitializePostgresConn()
	dbConn := db.PostgreSQL
	rbacRepo := repository.NewRBACRepository(dbConn, nil)
	rbac.TraversePermission(func(role rbac.Role, rsc rbac.Resource, act rbac.Action) {
		rra := model.RoleResourceAction{
			Role:     role,
			Resource: rsc,
			Action:   act,
		}
		err := rbacRepo.CreateRoleResourceAction(cmd.Context(), &rra)
		if err != nil {
			logrus.Error(err)
		}
	})
	logrus.Info("finished migrate rbac permissions")
}
