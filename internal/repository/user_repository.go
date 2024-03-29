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
	"time"
)

type userRepository struct {
	db           *gorm.DB
	cacheManager cacher.CacheManager
}

func NewUserRepository(db *gorm.DB, cacheManager cacher.CacheManager) model.UserRepository {
	return &userRepository{
		db:           db,
		cacheManager: cacheManager,
	}
}

func (u *userRepository) Create(ctx context.Context, userID int64, user *model.User) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":    utils.DumpIncomingContext(ctx),
		"userID": userID,
		"user":   utils.Dump(user),
	})
	user.CreatedBy = userID
	user.UpdatedBy = userID

	if err := u.db.WithContext(ctx).Create(user).Error; err != nil {
		logger.Error(err)
		return err
	}

	if err := u.deleteCache(user); err != nil {
		logger.Error(err)
	}

	return nil
}

func (u *userRepository) Update(ctx context.Context, userID int64, user *model.User) (*model.User, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":    utils.DumpIncomingContext(ctx),
		"userID": userID,
		"user":   utils.Dump(user),
	})

	user.UpdatedAt = time.Now()
	user.UpdatedBy = userID

	if err := u.db.WithContext(ctx).Updates(user).Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	if err := u.deleteCache(user); err != nil {
		logger.Error(err)
	}

	return u.FindByID(ctx, user.ID)
}

func (u *userRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": utils.DumpIncomingContext(ctx),
		"id":  id,
	})

	cacheKey := u.newCacheKeyByID(id)
	if !config.DisableCaching() {
		reply, mu, err := u.findFromCacheByKey(cacheKey)
		defer cacher.SafeUnlock(mu)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		if mu == nil {
			return reply, nil
		}
	}

	user := &model.User{}
	err := u.db.WithContext(ctx).Take(user, "id = ?", id).Error
	switch err {
	case nil:
	case gorm.ErrRecordNotFound:
		storeNil(u.cacheManager, cacheKey)
		return nil, nil
	default:
		logger.Error(err)
		return nil, err
	}

	err = u.cacheManager.StoreWithoutBlocking(cacher.NewItem(cacheKey, utils.Dump(user)))
	if err != nil {
		logger.Error(err)
	}

	return user, nil
}

func (u *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   utils.DumpIncomingContext(ctx),
		"email": email,
	})

	cacheKey := u.newCacheKeyByEmail(email)
	if !config.DisableCaching() {
		id, mu, err := u.findIDFromCacheByKey(cacheKey)
		defer cacher.SafeUnlock(mu)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		if mu == nil {
			return u.FindByID(ctx, id)
		}
	}

	var id int64
	err := u.db.WithContext(ctx).Model(model.User{}).Select("id").Take(&id, "email = ?", email).Error
	switch err {
	case nil:
		err := u.cacheManager.StoreWithoutBlocking(cacher.NewItem(cacheKey, id))
		if err != nil {
			logger.Error(err)
		}

		return u.FindByID(ctx, id)
	case gorm.ErrRecordNotFound:
		storeNil(u.cacheManager, cacheKey)
		return nil, nil
	default:
		logger.Error(err)
		return nil, err
	}
}

func (u *userRepository) IsLoginByEmailPasswordLocked(ctx context.Context, email string) (bool, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   utils.DumpIncomingContext(ctx),
		"email": email,
	})

	key := u.newLoginByEmailPasswordAttemptsCacheKeyByEmail(email)
	ttl, err := u.cacheManager.GetTTL(key)
	if err != nil {
		logger.Error(err)
		return false, err
	}

	loginAttempts, mu, err := u.findIntValueFromCacheByKey(key)
	defer cacher.SafeUnlock(mu)
	if err != nil {
		logger.Error(err)
		return false, err
	}

	if ttl > 0 && loginAttempts >= config.LoginRetryAttempts() {
		return true, nil
	}

	return false, nil
}

func (u *userRepository) IncrementLoginByEmailPasswordRetryAttempts(ctx context.Context, email string) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":   utils.DumpIncomingContext(ctx),
		"email": email,
	})

	key := u.newLoginByEmailPasswordAttemptsCacheKeyByEmail(email)
	if err := u.cacheManager.IncreaseCachedValueByOne(key); err != nil {
		logger.Error(err)
		return err
	}

	// resets the ttl duration everytime the attempts is incremented
	if err := u.cacheManager.Expire(key, config.LoginLockTTL()); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (u *userRepository) FindPasswordByID(ctx context.Context, id int64) ([]byte, error) {
	logger := logrus.WithFields(logrus.Fields{
		"ctx": utils.DumpIncomingContext(ctx),
		"id":  id,
	})

	cacheKey := u.newPasswordCacheKeyByID(id)
	if !config.DisableCaching() {
		reply, mu, err := u.findStringValueFromCacheByKey(cacheKey)
		defer cacher.SafeUnlock(mu)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		if mu == nil {
			return []byte(reply), nil
		}
	}

	var pass string
	err := u.db.WithContext(ctx).Model(model.User{}).Select("password").Take(&pass, "id = ?", id).Error
	switch err {
	case nil:
		err := u.cacheManager.StoreWithoutBlocking(cacher.NewItem(cacheKey, pass))
		if err != nil {
			logger.Error(err)
		}

		return []byte(pass), err
	case gorm.ErrRecordNotFound:
		return nil, nil
	default:
		logger.Error(err)
		return nil, err
	}
}

func (u *userRepository) UpdatePasswordByID(ctx context.Context, userID int64, password string) error {
	logger := logrus.WithFields(logrus.Fields{
		"ctx":    utils.DumpIncomingContext(ctx),
		"userID": userID,
	})

	updateFields := map[string]interface{}{
		"password": password,
	}
	err := u.db.WithContext(ctx).Model(model.User{}).Where("id = ?", userID).Updates(updateFields).Error
	if err != nil {
		logger.Error(err)
		return err
	}

	err = u.cacheManager.DeleteByKeys([]string{
		u.newPasswordCacheKeyByID(userID),
		u.newCacheKeyByID(userID),
	})
	if err != nil {
		logger.Error(err)
	}

	return nil
}

func (u *userRepository) deleteCache(user *model.User) error {
	return u.cacheManager.DeleteByKeys([]string{
		u.newCacheKeyByID(user.ID),
		u.newCacheKeyByEmail(user.Email),
		u.newPasswordCacheKeyByID(user.ID),
	})
}

func (u *userRepository) newCacheKeyByEmail(email string) string {
	return fmt.Sprintf("cache:id:user_email:%s", email)
}

func (u *userRepository) newCacheKeyByID(id int64) string {
	return fmt.Sprintf("cache:object:user:id:%d", id)
}

func (u *userRepository) newPasswordCacheKeyByID(id int64) string {
	return fmt.Sprintf("cache:password:id:%d", id)
}

func (u *userRepository) newLoginByEmailPasswordAttemptsCacheKeyByEmail(email string) string {
	return fmt.Sprintf("cache:login_attempts:email_password:user_email:%s", email)
}

func (u *userRepository) findFromCacheByKey(key string) (reply *model.User, mu *redsync.Mutex, err error) {
	var rep interface{}
	rep, mu, err = u.cacheManager.GetOrLock(key)
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

func (u *userRepository) findIDFromCacheByKey(key string) (reply int64, mu *redsync.Mutex, err error) {
	var rep interface{}
	rep, mu, err = u.cacheManager.GetOrLock(key)
	if err != nil || rep == nil {
		return
	}

	bt, _ := rep.([]byte)
	if bt == nil {
		return
	}

	reply = utils.StringToInt64(string(bt))
	return
}

func (u *userRepository) findIntValueFromCacheByKey(key string) (reply int, mu *redsync.Mutex, err error) {
	var rep interface{}
	rep, mu, err = u.cacheManager.GetOrLock(key)
	if err != nil || rep == nil {
		return
	}

	bt, _ := rep.([]byte)
	if bt == nil {
		return
	}

	reply = utils.StringToInt(string(bt))
	return
}

func (u *userRepository) findStringValueFromCacheByKey(key string) (reply string, mu *redsync.Mutex, err error) {
	var rep interface{}
	rep, mu, err = u.cacheManager.GetOrLock(key)
	if err != nil || rep == nil {
		return
	}

	bt, _ := rep.([]byte)
	if bt == nil {
		return
	}

	reply = string(bt)
	return
}
