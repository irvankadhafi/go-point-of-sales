package console

import (
	"context"
	"github.com/irvankadhafi/go-point-of-sales/cacher"
	"github.com/irvankadhafi/go-point-of-sales/internal/config"
	"github.com/irvankadhafi/go-point-of-sales/internal/db"
	"github.com/irvankadhafi/go-point-of-sales/internal/helper"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/internal/repository"
	"github.com/irvankadhafi/go-point-of-sales/rbac"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

var seedCmd = &cobra.Command{
	Use:   "seeder",
	Short: "run seed-user",
	Long:  `This subcommand seeding user`,
	Run:   seeder,
}

func init() {
	RootCmd.AddCommand(seedCmd)
}

func seeder(cmd *cobra.Command, args []string) {
	// Initiate all connection like db, redis, etc
	db.InitializePostgresConn()
	generalCacher := cacher.ConstructCacheManager()

	redisOpts := &db.RedisConnectionPoolOptions{
		DialTimeout:     config.RedisDialTimeout(),
		ReadTimeout:     config.RedisReadTimeout(),
		WriteTimeout:    config.RedisWriteTimeout(),
		IdleCount:       config.RedisMaxIdleConn(),
		PoolSize:        config.RedisMaxActiveConn(),
		IdleTimeout:     240 * time.Second,
		MaxConnLifetime: 1 * time.Minute,
	}

	redisConn, err := db.NewRedigoRedisConnectionPool(config.RedisCacheHost(), redisOpts)
	continueOrFatal(err)
	defer helper.WrapCloser(redisConn.Close)

	redisLockConn, err := db.NewRedigoRedisConnectionPool(config.RedisLockHost(), redisOpts)
	continueOrFatal(err)
	defer helper.WrapCloser(redisLockConn.Close)

	generalCacher.SetConnectionPool(redisConn)
	generalCacher.SetLockConnectionPool(redisLockConn)
	generalCacher.SetDefaultTTL(config.CacheTTL())

	userRepo := repository.NewUserRepository(db.PostgreSQL, generalCacher)
	appClientRepo := repository.NewAppClientRepository(db.PostgreSQL, generalCacher)

	userAdminCipherPwd, err := helper.HashString("123456")
	if err != nil {
		logrus.Error(err)
	}

	userAdmin := &model.User{
		ID:       utils.GenerateID(),
		Name:     "Irvan Kadhafi",
		Email:    "irvankadhafi@mail.com",
		Password: userAdminCipherPwd,
		Status:   model.StatusActive,
		Role:     rbac.RoleAdmin,
	}

	err = userRepo.Create(context.Background(), userAdmin.ID, userAdmin)
	if err != nil {
		return
	}

	userMemberCipherPwd, err := helper.HashString("123456")
	if err != nil {
		logrus.Error(err)
	}

	userCashier := &model.User{
		ID:       utils.GenerateID(),
		Name:     "John Doe",
		Email:    "johndoe@mail.com",
		Password: userMemberCipherPwd,
		Status:   model.StatusActive,
		Role:     rbac.RoleCashiers,
	}

	err = userRepo.Create(context.Background(), userCashier.ID, userCashier)
	if err != nil {
		return
	}

	appClientCMS := &model.AppClient{
		ID:           utils.GenerateID(),
		ClientID:     "cms-app",
		ClientSecret: "JgkDLJM64z9vKJPnLNcw3yH2M2tmmmBI",
	}

	err = appClientRepo.Create(context.Background(), appClientCMS)
	if err != nil {
		return
	}

	logrus.Warn("DONE!")
}
