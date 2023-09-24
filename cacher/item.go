package cacher

import (
	"time"
)

type (
	Item interface {
		GetTTLInt64() int64
		GetKey() string
		GetValue() any
		SetTTL(ttl time.Duration)
	}

	item struct {
		key   string
		value any
		ttl   time.Duration
	}
)

// WithTTL defines a custom time-to-live (TTL) for an item in the GetOrSet function.
func WithTTL(ttl time.Duration) func(Item) {
	return func(i Item) {
		i.SetTTL(ttl)
	}
}

// NewItem creates a new cache item with the given key and value.
func NewItem(key string, value any) Item {
	return &item{
		key:   key,
		value: value,
	}
}

// NewItemWithCustomTTL creates a new cache item with the given key, value, and custom TTL.
func NewItemWithCustomTTL(key string, value any, customTTL time.Duration) Item {
	return &item{
		key:   key,
		value: value,
		ttl:   customTTL,
	}
}

// GetTTLInt64 returns the TTL of the item in seconds as an int64 value.
func (item *item) GetTTLInt64() int64 {
	return int64(item.ttl.Seconds())
}

// SetTTL sets the time-to-live (TTL) for the item.
func (item *item) SetTTL(ttl time.Duration) {
	item.ttl = ttl
}

// GetKey returns the key of the item.
func (item *item) GetKey() string {
	return item.key
}

// GetValue returns the value of the item.
func (item *item) GetValue() any {
	return item.value
}
