package cache

import "time"

type Cache interface {
	Set(key string, value any)
	SetWithTTL(key string, value any, ttl time.Duration)
	Get(key string) (any, bool)
	Keys() []string
	Delete(key string)
	Clear()
}
