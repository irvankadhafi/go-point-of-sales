package repository

import (
	"github.com/irvankadhafi/go-point-of-sales/cacher"
	"github.com/irvankadhafi/go-point-of-sales/internal/config"
	"github.com/irvankadhafi/go-point-of-sales/internal/db"
	"github.com/irvankadhafi/go-point-of-sales/internal/model/mock"
	"os"
	"strconv"
	"testing"
	"time"

	"gorm.io/driver/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis"
	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	"github.com/golang/mock/gomock"
	redigo "github.com/gomodule/redigo/redis"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func initializeTest() {
	config.GetConf()
	setupLogger()
}

func setupLogger() {
	formatter := runtime.Formatter{
		ChildFormatter: &log.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		},
		Line: true,
		File: true,
	}

	log.SetFormatter(&formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)

	verbose, _ := strconv.ParseBool(os.Getenv("VERBOSE"))
	if verbose {
		log.SetLevel(log.DebugLevel)
	}
}

func initializeCockroachMockConn() (db *gorm.DB, mock sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	db, err = gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	return
}

type repoTestKit struct {
	miniredis       *miniredis.Miniredis
	dbmock          sqlmock.Sqlmock
	db              *gorm.DB
	cacheKeeper     cacher.CacheManager
	ctrl            *gomock.Controller
	mockAuditRepo   *mock.MockAuditRepository
	mockUserRepo    *mock.MockUserRepository
	mockSessionRepo *mock.MockSessionRepository
}

func initializeRepoTestKit(t *testing.T) (kit *repoTestKit, close func()) {
	mr, _ := miniredis.Run()
	r, err := newRedisConnPool("redis://" + mr.Addr())
	require.NoError(t, err)

	k := cacher.ConstructCacheManager()
	k.SetDisableCaching(false)
	k.SetConnectionPool(r)
	k.SetLockConnectionPool(r)
	k.SetWaitTime(1 * time.Second) // override wait time to 1 second

	dbconn, dbmock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: dbconn}), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	gormDB.Logger = db.NewGormCustomLogger()

	ctrl := gomock.NewController(t)
	auditRepo := mock.NewMockAuditRepository(ctrl)
	userRepo := mock.NewMockUserRepository(ctrl)
	sessionRepo := mock.NewMockSessionRepository(ctrl)

	tk := &repoTestKit{
		cacheKeeper:     k,
		miniredis:       mr,
		ctrl:            ctrl,
		dbmock:          dbmock,
		db:              gormDB,
		mockAuditRepo:   auditRepo,
		mockUserRepo:    userRepo,
		mockSessionRepo: sessionRepo,
	}

	close = func() {
		if conn, _ := tk.db.DB(); conn != nil {
			_ = conn.Close()
		}
		tk.miniredis.Close()
	}

	return tk, close
}

func newRedisConnPool(url string) (*redigo.Pool, error) {
	redisOpts := &db.RedisConnectionPoolOptions{
		DialTimeout:     config.RedisDialTimeout(),
		ReadTimeout:     config.RedisReadTimeout(),
		WriteTimeout:    config.RedisWriteTimeout(),
		IdleCount:       config.RedisMaxIdleConn(),
		PoolSize:        config.RedisMaxActiveConn(),
		IdleTimeout:     240 * time.Second,
		MaxConnLifetime: 1 * time.Minute,
	}

	c, err := db.NewRedigoRedisConnectionPool(url, redisOpts)
	if err != nil {
		return nil, err
	}

	return c, nil
}
