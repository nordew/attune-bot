package cache

import (
	"sync"
	"time"
)

type cachedItem struct {
	value      any
	expiration time.Time
}

type inMemoryCache struct {
	mu    sync.RWMutex
	items map[string]cachedItem
}

func NewCache() Cache {
	return &inMemoryCache{
		items: make(map[string]cachedItem),
	}
}

func (c *inMemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		c.Delete(key)
		return nil, false
	}

	return item.value, true
}

func (c *inMemoryCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

func (c *inMemoryCache) Set(key string, value any) {
	c.SetWithTTL(key, value, 0)
}

func (c *inMemoryCache) SetWithTTL(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}
	c.items[key] = cachedItem{
		value:      value,
		expiration: expiration,
	}
}

func (c *inMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *inMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]cachedItem)
}
