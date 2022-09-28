package cache

import (
	"sync"
)

type LocalCache struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func GetLocalCache() *LocalCache {
	lc := &LocalCache{
		data: make(map[string]interface{}),
	}

	return lc
}

func (c *LocalCache) Put(key string, value interface{}) {
	c.mu.Lock()
	c.data[key] = value
	c.mu.Unlock()
}

func (c *LocalCache) PutAll(entries map[string]interface{}) {
	for k, v := range entries {
		c.Put(k, v)
	}
}

func (c *LocalCache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	data, _ := c.data[key]
	c.mu.RUnlock()
	return data, nil
}

func (c *LocalCache) Has(key string) (bool, error) {
	c.mu.RLock()
	data, _ := c.data[key]
	c.mu.RUnlock()
	return data != nil, nil
}

func (c *LocalCache) GetAll(keys []string) map[string]interface{} {
	entries := make(map[string]interface{})
	for _, k := range keys {
		entries[k], _ = c.Get(k)
	}

	return entries
}

func (c *LocalCache) Clean(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

func (c *LocalCache) CleanAll() {
	c.mu.Lock()
	c.data = make(map[string]interface{})
	c.mu.Unlock()
}
