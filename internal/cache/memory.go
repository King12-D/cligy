package cache

import (
	"sync"
)

type MemoryCache struct {
	mu    sync.RWMutex
	store map[string][]byte
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		store: make(map[string][]byte),
	}
}

func (c *MemoryCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	data, ok := c.store[key]
	return data, ok
}

func (c *MemoryCache) Set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = data
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[string][]byte)
}
