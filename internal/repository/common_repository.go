package repository

import (
	"github.com/irvankadhafi/go-point-of-sales/cacher"
	"github.com/irvankadhafi/go-point-of-sales/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func storeNil(ck cacher.CacheManager, key string) {
	err := ck.StoreNil(key)
	if err != nil {
		logrus.Error(err)
	}
}

// scopeByPageAndLimit is a helper function to apply pagination on gorm query.
// it takes in 2 input as page and limit and returns a scope function
// that can be passed to gorm's db.Scopes method
// It is reusable to apply pagination on any query where it is needed
func scopeByPageAndLimit(page, limit int) func(d *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB { return db.Offset(utils.Offset(page, limit)).Limit(limit) }
}

// scopeMatchTSQuery is a helper function to apply plainto_tsquery on gorm query.
// it takes in a query string as input and returns a scope function
// that can be passed to gorm's db.Scopes method
// It is reusable to apply plainto_tsquery on any query where it is needed
func scopeMatchTSQuery(query string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name @@ plainto_tsquery(?)", query)
	}
}
