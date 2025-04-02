package cache

import (
	"context"
	"log"
	"time"
)

type StartCacheWorkerConfig struct {
	Cache    Cache
	Interval time.Duration
	StopCh   <-chan struct{}
}

func StartCacheWorker(
	ctx context.Context,
	cfg StartCacheWorkerConfig,
) {
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	log.Println("Cache worker started")
	for {
		select {
		case <-ctx.Done():
			log.Println("Cache worker context done")
		case <-ticker.C:
			cleanCache(cfg.Cache)
		case <-cfg.StopCh:
			log.Println("Cache worker stopped")
			return
		}
	}
}

func cleanCache(cache Cache) {
	memCache, ok := cache.(*inMemoryCache)
	if !ok {
		return
	}

	now := time.Now()

	memCache.mu.Lock()
	defer memCache.mu.Unlock()

	for key, item := range memCache.items {
		if !item.expiration.IsZero() && now.After(item.expiration) {
			delete(memCache.items, key)
		}
	}
}
