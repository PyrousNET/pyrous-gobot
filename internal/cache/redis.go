package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	conn *redis.Client
	ctx  context.Context
}

func GetRedisCache(connStr string) *RedisCache {
	uri, _ := url.Parse(connStr)
	password, _ := uri.User.Password()

	cch := &RedisCache{
		conn: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", uri.Hostname(), uri.Port()),
			Username: uri.User.Username(),
			Password: password,
		}),
	}

	cch.ctx = context.Background()

	return cch
}

func (rc *RedisCache) Put(key string, value interface{}) {
	jsonItem, _ := json.Marshal(value)
	if err := rc.conn.Set(rc.ctx, key, jsonItem, 0); err != nil {
		fmt.Println(err)
	}
}

func (rc *RedisCache) PutAll(entries map[string]interface{}) {
	for k, v := range entries {
		rc.Put(k, v)
	}
}

func (rc *RedisCache) Get(key string) (interface{}, bool, error) {
	value, err := rc.conn.Get(rc.ctx, key).Result()
	ok, _ := rc.conn.Exists(rc.ctx, key).Result()
	if err != nil {
		fmt.Println(err)
		return "", false, err
	}
	return value, ok > 0, err
}

func (rc *RedisCache) GetAll(keys []string) map[string]interface{} {
	entries := make(map[string]interface{})
	for _, k := range keys {
		entries[k], _, _ = rc.Get(k)
	}

	return entries
}

func (rc *RedisCache) Clean(key string) {
	if err := rc.conn.Del(rc.ctx, key); err != nil {
		fmt.Println(err)
	}
}

// CleanAll cleans the entire cache.
func (rc *RedisCache) CleanAll() {
	rc.conn.FlushDB(rc.ctx)
}

func (rc *RedisCache) GetKeys(prefix string) []string {
	return rc.conn.Keys(rc.ctx, prefix+"_*").Val()
}
