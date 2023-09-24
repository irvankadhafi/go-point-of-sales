package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/irvankadhafi/go-point-of-sales/cacher"
	"github.com/irvankadhafi/go-point-of-sales/internal/config"
	"github.com/irvankadhafi/go-point-of-sales/internal/model"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type appClientRepo struct {
	db          *gorm.DB
	cacheKeeper cacher.CacheManager
}

// NewAppClientRepository constructor
func NewAppClientRepository(
	db *gorm.DB,
	cacheKeeper cacher.CacheManager,
) model.AppClientRepository {
	return &appClientRepo{
		db:          db,
		cacheKeeper: cacheKeeper,
	}
}

// FindByClientID :nodoc:
func (s *appClientRepo) FindByClientID(ctx context.Context, clientID string) (*model.AppClient, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":      utils.DumpIncomingContext(ctx),
		"clientID": clientID,
	})

	cacheKey := s.newCacheKeyByClientID(clientID)
	if !config.DisableCaching() {
		reply, mu, err := s.findFromCacheByKey(cacheKey)
		defer cacher.SafeUnlock(mu)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		if mu == nil {
			return reply, nil
		}
	}

	appClient := model.AppClient{}
	err := s.db.WithContext(ctx).Take(&appClient, "client_id = ?", clientID).Error
	switch err {
	case nil:
		err := s.cacheKeeper.StoreWithoutBlocking(cacher.NewItem(cacheKey, utils.Dump(appClient)))
		if err != nil {
			logger.Error(err)
		}
		return &appClient, nil
	default:
		logger.Error(err)
		return nil, err
	}
}

// FindByID :nodoc:
func (s *appClientRepo) FindByID(ctx context.Context, id int64) (*model.AppClient, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": utils.DumpIncomingContext(ctx),
		"id":  id,
	})

	cacheKey := s.newCacheKeyByID(id)
	if !config.DisableCaching() {
		reply, mu, err := s.findFromCacheByKey(cacheKey)
		defer cacher.SafeUnlock(mu)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		if mu == nil {
			return reply, nil
		}
	}

	appClient := model.AppClient{}
	err := s.db.WithContext(ctx).Take(&appClient, "id = ?", id).Error
	switch err {
	case nil:
		err := s.cacheKeeper.StoreWithoutBlocking(cacher.NewItem(cacheKey, utils.Dump(appClient)))
		if err != nil {
			logger.Error(err)
		}
		return &appClient, nil
	default:
		logger.Error(err)
		return nil, err
	}
}

// Create :nodoc:
func (s *appClientRepo) Create(ctx context.Context, appClient *model.AppClient) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":       utils.DumpIncomingContext(ctx),
		"appClient": utils.Dump(appClient),
	})

	if err := s.db.WithContext(ctx).Create(appClient).Error; err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (s *appClientRepo) newCacheKeyByClientID(clientID string) string {
	return fmt.Sprintf("cache:object:client_id:%s", clientID)
}

func (s *appClientRepo) newCacheKeyByID(id int64) string {
	return fmt.Sprintf("cache:object:app_client:id:%d", id)
}

func (s *appClientRepo) findFromCacheByKey(key string) (reply *model.AppClient, mu *redsync.Mutex, err error) {
	var rep interface{}
	rep, mu, err = s.cacheKeeper.GetOrLock(key)
	if err != nil || rep == nil {
		return
	}

	bt, _ := rep.([]byte)
	if bt == nil {
		return
	}

	if err = json.Unmarshal(bt, &reply); err != nil {
		return
	}

	return
}
