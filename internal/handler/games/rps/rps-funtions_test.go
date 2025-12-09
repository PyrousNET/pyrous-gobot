package rps

import (
	"testing"
	"time"
)

type fakeCache struct {
	store      map[string]interface{}
	expiredKey string
	expiredTTL time.Duration
}

func newFakeCache() *fakeCache {
	return &fakeCache{store: make(map[string]interface{})}
}

func (c *fakeCache) Put(key string, value interface{}) {
	c.store[key] = value
}

func (c *fakeCache) PutAll(entries map[string]interface{}) {
	for k, v := range entries {
		c.Put(k, v)
	}
}

func (c *fakeCache) Get(key string) (interface{}, bool, error) {
	v, ok := c.store[key]
	return v, ok, nil
}

func (c *fakeCache) GetAll(keys []string) map[string]interface{} {
	out := make(map[string]interface{})
	for _, k := range keys {
		if v, ok := c.store[k]; ok {
			out[k] = v
		}
	}
	return out
}

func (c *fakeCache) Clean(key string) {
	delete(c.store, key)
}

func (c *fakeCache) GetKeys(prefix string) ([]string, error) {
	return nil, nil
}

func (c *fakeCache) CleanAll() {
	c.store = make(map[string]interface{})
}

func (c *fakeCache) Expire(key string, ttl time.Duration) {
	c.expiredKey = key
	c.expiredTTL = ttl
}

func TestUpdateRpsSetsTTL(t *testing.T) {
	cache := newFakeCache()
	player := RPS{Name: "alice", RpsPlaying: "game123", Rps: "rock"}

	_, err := UpdateRps(player, "chan1", cache)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedKey := "rps-alice-chan1"
	if cache.expiredKey != expectedKey {
		t.Fatalf("expire called with key %q, want %q", cache.expiredKey, expectedKey)
	}
	if cache.expiredTTL != rpsCacheTTL {
		t.Fatalf("expire called with ttl %s, want %s", cache.expiredTTL, rpsCacheTTL)
	}
}

func TestSetGameTTL(t *testing.T) {
	cache := newFakeCache()

	SetGameTTL(cache, "game-uuid")

	if cache.expiredKey != "game-uuid" {
		t.Fatalf("expire called with key %q, want %q", cache.expiredKey, "game-uuid")
	}
	if cache.expiredTTL != rpsCacheTTL {
		t.Fatalf("expire called with ttl %s, want %s", cache.expiredTTL, rpsCacheTTL)
	}
}
