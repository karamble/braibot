package fmp

import (
	"sync"
	"time"
)

// memCache is a process-wide TTL cache shared across all callers: a
// fundamentals bundle assembled for one user serves every later request
// for the same symbol within the window without further FMP calls.
type memCache struct {
	mu      sync.Mutex
	ttl     time.Duration
	entries map[string]cacheEntry
}

type cacheEntry struct {
	value   interface{}
	expires time.Time
}

func newMemCache(ttl time.Duration) *memCache {
	return &memCache{ttl: ttl, entries: make(map[string]cacheEntry)}
}

func (c *memCache) get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expires) {
		delete(c.entries, key)
		return nil, false
	}
	return e.value, true
}

func (c *memCache) put(key string, v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Lazy sweep: drop expired entries while we hold the lock. The map
	// stays small (one entry per symbol/query), so a full pass is cheap.
	now := time.Now()
	for k, e := range c.entries {
		if now.After(e.expires) {
			delete(c.entries, k)
		}
	}
	c.entries[key] = cacheEntry{value: v, expires: now.Add(c.ttl)}
}
