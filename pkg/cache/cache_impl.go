package cache

import "sync"

type inMemoryCache struct {
	mu    sync.RWMutex
	items map[string]any
}

func NewCache() Cache {
	return &inMemoryCache{
		items: make(map[string]any),
	}
}

func (c *inMemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.items[key]

	return val, ok
}

func (c *inMemoryCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = value
}

func (c *inMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

func (c *inMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]any)
}
