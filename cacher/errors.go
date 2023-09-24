package cacher

import "errors"

var (
	ErrWaitTooLong             = errors.New("wait too long")
	ErrKeyNotExist             = errors.New("key not exist")
	ErrInvalidCacheValue       = errors.New("invalid cache value")
	ErrFailedCastMultiResponse = errors.New("failed to cast cache multi response")
)
