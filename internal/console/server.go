package console

import (
	"context"
	"errors"
	"fmt"
	"github.com/irvankadhafi/go-point-of-sales/auth"
	"github.com/irvankadhafi/go-point-of-sales/cacher"
	_ "github.com/irvankadhafi/go-point-of-sales/docs"
	"github.com/irvankadhafi/go-point-of-sales/internal/config"
	"github.com/irvankadhafi/go-point-of-sales/internal/db"
	"github.com/irvankadhafi/go-point-of-sales/internal/delivery/httpsvc"
	"github.com/irvankadhafi/go-point-of-sales/internal/helper"
	"github.com/irvankadhafi/go-point-of-sales/internal/repository"
	"github.com/irvankadhafi/go-point-of-sales/internal/usecase"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	echoSwagger "github.com/swaggo/echo-swagger"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var runCmd = &cobra.Command{
	Use:   "server",
	Short: "run server",
	Long:  `This subcommand start the server`,
	Run:   run,
}

func init() {
	RootCmd.AddCommand(runCmd)
}

func run(cmd *cobra.Command, args []string) {
	// Initiate all connection like db, redis, etc
	db.InitializePostgresConn()
	authenticationCacher := cacher.ConstructCacheManager()
	generalCacher := cacher.ConstructCacheManager()
	pgDB, err := db.PostgreSQL.DB()
	continueOrFatal(err)
	defer helper.WrapCloser(pgDB.Close)

	redisOpts := &db.RedisConnectionPoolOptions{
		DialTimeout:     config.RedisDialTimeout(),
		ReadTimeout:     config.RedisReadTimeout(),
		WriteTimeout:    config.RedisWriteTimeout(),
		IdleCount:       config.RedisMaxIdleConn(),
		PoolSize:        config.RedisMaxActiveConn(),
		IdleTimeout:     240 * time.Second,
		MaxConnLifetime: 1 * time.Minute,
	}

	authRedisConn, err := db.NewRedigoRedisConnectionPool(config.RedisAuthCacheHost(), redisOpts)
	continueOrFatal(err)
	defer helper.WrapCloser(authRedisConn.Close)

	authRedisLockConn, err := db.NewRedigoRedisConnectionPool(config.RedisAuthCacheLockHost(), redisOpts)
	continueOrFatal(err)
	defer helper.WrapCloser(authRedisLockConn.Close)

	authenticationCacher.SetConnectionPool(authRedisConn)
	authenticationCacher.SetLockConnectionPool(authRedisLockConn)
	authenticationCacher.SetDefaultTTL(config.CacheTTL())

	redisConn, err := db.NewRedigoRedisConnectionPool(config.RedisCacheHost(), redisOpts)
	continueOrFatal(err)
	defer helper.WrapCloser(redisConn.Close)

	redisLockConn, err := db.NewRedigoRedisConnectionPool(config.RedisLockHost(), redisOpts)
	continueOrFatal(err)
	defer helper.WrapCloser(redisLockConn.Close)

	generalCacher.SetConnectionPool(redisConn)
	generalCacher.SetLockConnectionPool(redisLockConn)
	generalCacher.SetDefaultTTL(config.CacheTTL())

	auditRepo := repository.NewAuditRepository()
	rbacRepo := repository.NewRBACRepository(db.PostgreSQL, authenticationCacher)
	userRepo := repository.NewUserRepository(db.PostgreSQL, generalCacher)
	sessionRepo := repository.NewSessionRepository(db.PostgreSQL, authenticationCacher, userRepo)
	appClientRepo := repository.NewAppClientRepository(db.PostgreSQL, authenticationCacher)
	productRepo := repository.NewProductRepository(db.PostgreSQL, generalCacher, auditRepo)
	transactionDetailRepo := repository.NewTransactionDetailRepository(db.PostgreSQL, generalCacher)
	transactionRepo := repository.NewTransactionRepository(db.PostgreSQL, generalCacher, transactionDetailRepo, auditRepo)

	userUsecase := usecase.NewUserUsecase(userRepo)
	authUsecase := usecase.NewAuthUsecase(userRepo, sessionRepo, rbacRepo)
	userAuther := usecase.NewUserAutherAdapter(authUsecase)
	appClientUsecase := usecase.NewAppClientUsecase(appClientRepo)
	productUsecase := usecase.NewProductUsecase(productRepo)
	transactionUsecase := usecase.NewTransactionUsecase(transactionRepo, productRepo)

	httpServer := echo.New()
	httpMiddleware := auth.NewAuthenticationMiddleware(userAuther, authenticationCacher)

	httpServer.Pre(middleware.AddTrailingSlash())
	httpServer.Use(middleware.Logger())
	httpServer.Use(middleware.Recover())
	httpServer.Use(middleware.CORS())

	httpServer.GET("/swagger/*", echoSwagger.WrapHandler, middleware.RemoveTrailingSlash())
	apiGroup := httpServer.Group("/api")
	httpsvc.RouteService(apiGroup, authUsecase, userUsecase, appClientUsecase, productUsecase, transactionUsecase, httpMiddleware)

	sigCh := make(chan os.Signal, 1)
	errCh := make(chan error, 1)
	quitCh := make(chan bool, 1)
	signal.Notify(sigCh, os.Interrupt)

	go func() {
		for {
			select {
			case <-sigCh:
				gracefulShutdown(httpServer)
				quitCh <- true
			case e := <-errCh:
				log.Error(e)
				gracefulShutdown(httpServer)
				quitCh <- true
			}
		}
	}()

	go func() {
		// Start HTTP server
		if err := httpServer.Start(fmt.Sprintf(":%s", config.Port())); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	<-quitCh
	log.Info("exiting")
}

func gracefulShutdown(httpSvr *echo.Echo) {
	db.StopTickerCh <- true

	if httpSvr != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpSvr.Shutdown(ctx); err != nil {
			httpSvr.Logger.Fatal(err)
		}
	}
}
