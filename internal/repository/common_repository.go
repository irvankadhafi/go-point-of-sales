package repository

import (
	"github.com/irvankadhafi/go-point-of-sales/cacher"
	"github.com/sirupsen/logrus"
)

func storeNil(ck cacher.CacheManager, key string) {
	err := ck.StoreNil(key)
	if err != nil {
		logrus.Error(err)
	}
}
