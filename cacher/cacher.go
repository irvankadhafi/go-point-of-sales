package cacher

import (
	"encoding/json"
	"fmt"
	"github.com/irvankadhafi/go-point-of-sales/internal/config"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"time"

	"github.com/go-redsync/redsync/v4"
	redigosync "github.com/go-redsync/redsync/v4/redis/redigo"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/jpillora/backoff"
)

var nilValue = []byte("null")

type (
	GetterFn func() (any, error)

	CacheManager interface {
		Get(key string) (any, error)
		GetOrLock(key string) (any, *redsync.Mutex, error)
		GetOrSet(key string, fn GetterFn, opts ...func(Item)) ([]byte, error)

		// HASH BUCKET
		GetHashMemberOrLock(identifier string, key string) (any, *redsync.Mutex, error)
		GetHashMember(identifier string, key string) (any, error)
		StoreHashMember(identifier string, c Item) (err error)

		Store(*redsync.Mutex, Item) error
		StoreWithoutBlocking(Item) error
		StoreMultiWithoutBlocking([]Item) error
		StoreMultiPersist([]Item) error
		StoreNil(cacheKey string) error
		StoreNilWithCustomTTL(cacheKey string, customTTL time.Duration) error

		IncreaseCachedValueByOne(key string) error
		GetTTL(string) (int64, error)

		Expire(string, time.Duration) error
		ExpireMulti(map[string]time.Duration) error
		Purge(string) error
		DeleteByKeys(keys []string) error
		SetCachePrefix(string, string)

		CheckKeyExist(key string) (value bool, err error)

		AcquireLock(string) (*redsync.Mutex, error)
		SetDefaultTTL(time.Duration)
		SetNilTTL(time.Duration)
		SetConnectionPool(*redigo.Pool)
		SetLockConnectionPool(*redigo.Pool)
		SetLockDuration(time.Duration)
		SetLockTries(int)
		SetWaitTime(time.Duration)
		SetDisableCaching(bool)
	}

	cacheManager struct {
		environment    string
		prefixCacheKey string

		connPool       *redigo.Pool
		nilTTL         time.Duration
		defaultTTL     time.Duration
		waitTime       time.Duration
		disableCaching bool

		lockConnPool *redigo.Pool
		lockDuration time.Duration
		lockTries    int
	}

	itemWithKey struct {
		Key  string
		Item any
	}

	mutexWithKey struct {
		Key   string
		Mutex *redsync.Mutex
	}

	errorWithKey struct {
		key        string
		innerError error
	}
)

// Error implements built-in error interfaces
func (errorKey *errorWithKey) Error() string {
	var msg string
	if errorKey.innerError != nil {
		msg = errorKey.innerError.Error()
	}

	return fmt.Sprintf("err on key %s : %s", errorKey.key, msg)
}

// ConstructCacheManager is used to create an instance of CacheManager with default configuration.
func ConstructCacheManager() CacheManager {
	return &cacheManager{
		defaultTTL:     defaultTTL,
		nilTTL:         defaultNilTTL,
		prefixCacheKey: defaultPrefixCacheKey,
		environment:    config.Env(),
		lockDuration:   defaultLockDuration,
		lockTries:      defaultLockTries,
		waitTime:       defaultWaitTime,
		disableCaching: false,
	}
}

// Get is used to retrieve an item stored in the cache based on the key.
func (cache *cacheManager) Get(key string) (cachedItem any, err error) {
	if cache.disableCaching {
		return
	}

	cachedItem, err = get(cache.connPool.Get(), key)
	if err != nil && err != ErrKeyNotExist && err != redigo.ErrNil || cachedItem != nil {
		return
	}

	return nil, nil
}

// GetOrLock is used to retrieve an item from the cache based on the key. If the item is not found,
// it will acquire a lock and wait for the item to be available in the cache.
func (cache *cacheManager) GetOrLock(key string) (cachedItem any, mutex *redsync.Mutex, err error) {
	if cache.disableCaching {
		return
	}

	cachedItem, err = get(cache.connPool.Get(), key)
	if err != nil && err != ErrKeyNotExist && err != redigo.ErrNil || cachedItem != nil {
		return
	}

	mutex, err = cache.AcquireLock(key)
	if err == nil {
		return
	}

	startTime := time.Now()
	for {
		backoffRetries := &backoff.Backoff{
			Min:    20 * time.Millisecond,
			Max:    200 * time.Millisecond,
			Jitter: true,
		}

		if !cache.isLocked(key) {
			cachedItem, err = get(cache.connPool.Get(), key)
			if err != nil {
				if err == ErrKeyNotExist {
					mutex, err = cache.AcquireLock(key)
					if err == nil {
						return nil, mutex, nil
					}

					goto Wait
				}
				return nil, nil, err
			}
			return cachedItem, nil, nil
		}
	Wait:
		elapsed := time.Since(startTime)
		if elapsed >= cache.waitTime {
			break
		}

		time.Sleep(backoffRetries.Duration())
	}

	return nil, nil, ErrWaitTooLong
}

// GetOrSet is used to retrieve a value from the cache based on a given key.
// If the value is not found in the cache, it will be fetched using a getter function and then stored in the cache for future use.
// The function also provides options for customizing the caching behavior through optional functional parameters opts
func (cache *cacheManager) GetOrSet(key string, fn GetterFn, opts ...func(Item)) (res []byte, err error) {
	if cache.disableCaching {
		myResp, err := fn()
		if err != nil {
			return nil, err
		}

		return json.Marshal(myResp)
	}

	cachedValue, mu, err := cache.GetOrLock(key)
	if err != nil {
		return
	}
	if cachedValue != nil {
		res, ok := cachedValue.([]byte)
		if !ok {
			return nil, ErrInvalidCacheValue
		}

		return res, nil
	}

	// handle if nil value is cached
	if mu == nil {
		return
	}

	defer SafeUnlock(mu)
	item, err := fn()
	if err != nil {
		return
	}

	if item == nil {
		_ = cache.StoreNil(key)
		return
	}

	cachedValue, err = json.Marshal(item)
	if err != nil {
		return
	}

	cacheItem := NewItem(key, cachedValue)
	for _, o := range opts {
		o(cacheItem)
	}
	_ = cache.Store(mu, cacheItem)
	return cachedValue.([]byte), nil
}

// GetHashMemberOrLock :nodoc:
func (cache *cacheManager) GetHashMemberOrLock(identifier string, key string) (cachedItem any, mutex *redsync.Mutex, err error) {
	if cache.disableCaching {
		return
	}

	lockKey := fmt.Sprintf("%s:%s", identifier, key)

	cachedItem, err = cache.GetHashMember(identifier, key)
	if err != nil && err != redigo.ErrNil && err != ErrKeyNotExist || cachedItem != nil {
		return
	}

	mutex, err = cache.AcquireLock(lockKey)
	if err == nil {
		return // nolint:nilerr
	}

	start := time.Now()
	for {
		b := &backoff.Backoff{
			Min:    20 * time.Millisecond,
			Max:    200 * time.Millisecond,
			Jitter: true,
		}

		if !cache.isLocked(lockKey) {
			cachedItem, err = cache.GetHashMember(identifier, key)
			if err != nil {
				if err == ErrKeyNotExist {
					mutex, err = cache.AcquireLock(lockKey)
					if err == nil {
						return nil, mutex, nil
					}

					goto Wait
				}
				return nil, nil, err
			}
			return cachedItem, nil, nil
		}

	Wait:
		elapsed := time.Since(start)
		if elapsed >= cache.waitTime {
			break
		}

		time.Sleep(b.Duration())
	}

	return nil, nil, ErrWaitTooLong
}

// GetHashMember :nodoc:
func (cache *cacheManager) GetHashMember(identifier string, key string) (value any, err error) {
	if cache.disableCaching {
		return
	}

	client := cache.connPool.Get()
	defer func() {
		_ = client.Close()
	}()

	return getHashMember(client, identifier, key)
}

// StoreHashMember :nodoc:
func (cache *cacheManager) StoreHashMember(identifier string, c Item) (err error) {
	if cache.disableCaching {
		return nil
	}

	client := cache.connPool.Get()
	defer func() {
		_ = client.Close()
	}()

	err = client.Send("MULTI")
	if err != nil {
		return err
	}
	_, err = client.Do("HSET", identifier, c.GetKey(), c.GetValue())
	if err != nil {
		return err
	}
	_, err = client.Do("EXPIRE", identifier, cache.decideCacheTTL(c))
	if err != nil {
		return err
	}

	_, err = client.Do("EXEC")
	return
}

// Store is used to store an item in the cache with an optional mutex lock.
func (cache *cacheManager) Store(mutex *redsync.Mutex, item Item) error {
	if cache.disableCaching {
		return nil
	}
	defer SafeUnlock(mutex)

	client := cache.connPool.Get()
	defer utils.WrapCloser(client.Close)

	_, err := client.Do("SETEX", item.GetKey(), cache.decideCacheTTL(item), item.GetValue())
	return err
}

// StoreWithoutBlocking is used to store an item in the cache without acquiring a lock.
func (cache *cacheManager) StoreWithoutBlocking(item Item) error {
	if cache.disableCaching {
		return nil
	}

	client := cache.connPool.Get()
	defer utils.WrapCloser(client.Close)

	_, err := client.Do("SETEX", item.GetKey(), cache.decideCacheTTL(item), item.GetValue())
	return err
}

// StoreMultiWithoutBlocking is used to store multiple items in the cache without acquiring locks.
func (cache *cacheManager) StoreMultiWithoutBlocking(items []Item) error {
	if cache.disableCaching {
		return nil
	}

	client := cache.connPool.Get()
	defer utils.WrapCloser(client.Close)

	if err := client.Send("MULTI"); err != nil {
		return err
	}

	for _, item := range items {
		if err := client.Send("SETEX", item.GetKey(), cache.decideCacheTTL(item), item.GetValue()); err != nil {
			return err
		}
	}

	_, err := client.Do("EXEC")
	return err
}

// StoreMultiPersist is used to store multiple items in the cache and persist them indefinitely.
func (cache *cacheManager) StoreMultiPersist(items []Item) error {
	if cache.disableCaching {
		return nil
	}

	client := cache.connPool.Get()
	defer utils.WrapCloser(client.Close)

	if err := client.Send("MULTI"); err != nil {
		return err
	}

	for _, item := range items {
		if err := client.Send("SET", item.GetKey(), item.GetValue()); err != nil {
			return err
		}

		if err := client.Send("PERSIST", item.GetKey()); err != nil {
			return err
		}
	}

	_, err := client.Do("EXEC")
	return err
}

// StoreNil is used to store a nil value in the cache with a default time-to-live (TTL).
func (cache *cacheManager) StoreNil(cacheKey string) error {
	item := NewItemWithCustomTTL(cacheKey, nilValue, cache.nilTTL)

	return cache.StoreWithoutBlocking(item)
}

// StoreNilWithCustomTTL is used to store a nil value in the cache with a custom time-to-live (TTL).
func (cache *cacheManager) StoreNilWithCustomTTL(cacheKey string, customTTL time.Duration) error {
	item := NewItemWithCustomTTL(cacheKey, nilValue, customTTL)

	return cache.StoreWithoutBlocking(item)
}

// IncreaseCachedValueByOne will increments the number stored at key by one.
// If the key does not exist, it is set to 0 before performing the operation
func (cache *cacheManager) IncreaseCachedValueByOne(key string) error {
	if cache.disableCaching {
		return nil
	}

	client := cache.connPool.Get()
	defer func() {
		_ = client.Close()
	}()

	_, err := client.Do("INCR", key)
	return err
}

// Expire is used to set an expiration time for a cache item based on the key.
func (cache *cacheManager) Expire(key string, duration time.Duration) (err error) {
	if cache.disableCaching {
		return nil
	}

	client := cache.connPool.Get()
	defer utils.WrapCloser(client.Close)

	_, err = client.Do("EXPIRE", key, int64(duration.Seconds()))
	return
}

func (cache *cacheManager) GetTTL(name string) (value int64, err error) {
	client := cache.connPool.Get()
	defer func() {
		_ = client.Close()
	}()

	val, err := client.Do("TTL", name)
	if err != nil {
		return
	}

	value = val.(int64)
	return
}

// ExpireMulti is used to set expiration times for multiple cache items based on their keys.
func (cache *cacheManager) ExpireMulti(items map[string]time.Duration) error {
	if cache.disableCaching {
		return nil
	}

	client := cache.connPool.Get()
	defer utils.WrapCloser(client.Close)

	if err := client.Send("MULTI"); err != nil {
		return err
	}

	for key, duration := range items {
		if err := client.Send("EXPIRE", key, int64(duration.Seconds())); err != nil {
			return err
		}
	}

	_, err := client.Do("EXEC")
	return err
}

// Purge is used to remove all cache items that match a given pattern.
func (cache *cacheManager) Purge(matchString string) error {
	if cache.disableCaching {
		return nil
	}

	client := cache.connPool.Get()
	defer utils.WrapCloser(client.Close)

	var cursor any
	var stop []uint8
	cursor = "0"
	delCount := 0
	for {
		res, err := redigo.Values(client.Do("SCAN", cursor, "MATCH", matchString, "COUNT", 500000))
		if err != nil {
			return err
		}

		stop = res[0].([]uint8)
		if foundKeys, ok := res[1].([]any); ok {
			if len(foundKeys) > 0 {
				err = client.Send("DEL", foundKeys...)
				if err != nil {
					return err
				}

				delCount++
			}

			// ascii for '0' is 48
			if stop[0] == 48 {
				break
			}
		}

		cursor = res[0]
	}

	if delCount > 0 {
		_ = client.Flush()
	}

	return nil
}

// DeleteByKeys is used to delete cache items based on their keys.
func (cache *cacheManager) DeleteByKeys(keys []string) error {
	if cache.disableCaching {
		return nil
	}

	if len(keys) <= 0 {
		return nil
	}

	client := cache.connPool.Get()
	defer utils.WrapCloser(client.Close)

	var redisKeys []any
	for _, key := range keys {
		redisKeys = append(redisKeys, key)
	}

	_, err := client.Do("DEL", redisKeys...)
	return err
}

// AcquireLock is used to acquire a lock on a cache item based on the key.
func (cache *cacheManager) AcquireLock(key string) (*redsync.Mutex, error) {
	pool := redigosync.NewPool(cache.lockConnPool)

	mutex := redsync.New(pool).NewMutex(
		"lock:"+key,
		redsync.WithExpiry(cache.lockDuration),
		redsync.WithTries(cache.lockTries),
	)

	return mutex, mutex.Lock()
}

// SetNilTTL is used to set the time-to-live (TTL) for nil values stored in the cache.
func (cache *cacheManager) SetNilTTL(duration time.Duration) {
	cache.nilTTL = duration
}

// SetConnectionPool is used to set the connection pool for the cache manager.
func (cache *cacheManager) SetConnectionPool(pool *redigo.Pool) {
	cache.connPool = pool

	//	by default, lock connection pool use same connection with primary default connection
	cache.lockConnPool = pool
}

// SetLockConnectionPool is used to set the connection pool for lock acquisition in the cache manager.
func (cache *cacheManager) SetLockConnectionPool(pool *redigo.Pool) {
	cache.lockConnPool = pool
}

// SetLockDuration is used to set the lock duration for cache items in the cache manager.
func (cache *cacheManager) SetLockDuration(duration time.Duration) {
	cache.lockDuration = duration
}

// SetLockTries is used to set the number of lock acquisition tries in the cache manager.
func (cache *cacheManager) SetLockTries(lockTries int) {
	cache.lockTries = lockTries
}

// SetWaitTime is used to set the maximum wait time for acquiring a lock in the cache manager.
func (cache *cacheManager) SetWaitTime(duration time.Duration) {
	cache.waitTime = duration
}

// SetDisableCaching is used to enable or disable caching in the cache manager.
func (cache *cacheManager) SetDisableCaching(disableCaching bool) {
	cache.disableCaching = disableCaching
}

// SetCachePrefix is used to set the cache key prefix and environment in the cache manager.
func (cache *cacheManager) SetCachePrefix(prefix, env string) {
	cache.prefixCacheKey = prefix
	cache.environment = env
}

// SetDefaultTTL is used to set the default time-to-live (TTL) for cache items in the cache manager.
func (cache *cacheManager) SetDefaultTTL(duration time.Duration) {
	cache.defaultTTL = duration
}

// CheckKeyExist is used to check if a cache key exists.
func (cache *cacheManager) CheckKeyExist(key string) (value bool, err error) {
	client := cache.connPool.Get()
	defer utils.WrapCloser(client.Close)

	val, err := client.Do("EXISTS", key)
	res, ok := val.(int64)
	if ok && res > 0 {
		value = true
	}

	return
}

// decideCacheTTL is used to determine the time-to-live (TTL) for a cache item.
func (cache *cacheManager) decideCacheTTL(c Item) (ttl int64) {
	if ttl = c.GetTTLInt64(); ttl > 0 {
		return
	}

	return int64(cache.defaultTTL.Seconds())
}

// isLocked is used to check if a cache item is locked.
func (cache *cacheManager) isLocked(key string) bool {
	client := cache.lockConnPool.Get()
	defer utils.WrapCloser(client.Close)

	reply, err := client.Do("GET", "lock:"+key)
	if err != nil || reply == nil {
		return false
	}

	return true
}
