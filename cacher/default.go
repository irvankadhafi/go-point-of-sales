package cacher

import "time"

const (
	defaultTTL            = 15 * time.Minute
	defaultNilTTL         = 5 * time.Minute
	defaultLockDuration   = 1 * time.Minute
	defaultLockTries      = 1
	defaultWaitTime       = 15 * time.Second
	defaultPrefixCacheKey = "customer-service"
)
