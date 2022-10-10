package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/go-redis/redis/v8"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/nitishm/go-rejson/v4"
	"github.com/nitishm/go-rejson/v4/rjs"
)

type RedisCache struct {
	conn     *redis.Client
	ctx      context.Context
	jHandler *rejson.Handler
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

	rh := rejson.NewReJSONHandler()
	rh.SetGoRedisClient(cch.conn)
	cch.jHandler = rh

	return cch
}

func (rc *RedisCache) Put(key string, value interface{}) {
	if err := rc.conn.Set(rc.ctx, key, value, 0); err != nil {
		fmt.Println(err)
	}
}

func (rc *RedisCache) PutAll(entries map[string]interface{}) {
	for k, v := range entries {
		rc.Put(k, v)
	}
}

func (rc *RedisCache) Get(key string) (interface{}, bool, error) {
	numkeys, _ := rc.conn.Exists(rc.ctx, key).Result()
	if ok := numkeys > 0; ok {
		value, err := rc.conn.Get(rc.ctx, key).Result()
		if err != nil {
			return "", false, err
		}
		return value, ok, nil
	}

	return "", false, nil
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

func (rc *RedisCache) GetKeys(prefix string) ([]string, error) {
	return rc.conn.Keys(rc.ctx, prefix+"*").Result()
}

func (rc *RedisCache) GetJsonObj(key string) ([]byte, error) {
	return redigo.Bytes(
		rc.jHandler.JSONGet(key, ".",
			rjs.GETOptionINDENT,
			rjs.GETOptionNOESCAPE,
			rjs.GETOptionNEWLINE,
			rjs.GETOptionSPACE,
		),
	)
}

func (rc *RedisCache) GetJsonObjKeys(key string) (interface{}, error) {
	return rc.jHandler.JSONObjKeys(key, ".")
}

func (rc *RedisCache) PutJsonObj(key string, object interface{}) error {
	_, err := rc.jHandler.JSONSet(key, ".", object)

	return err
}

func (rc *RedisCache) PutJsonFromLocalFile() error {
	jf, err := os.Open("./local.json")
	if err != nil {
		return err
	}
	defer jf.Close()
	b, _ := ioutil.ReadAll(jf)

	var m map[string]interface{}
	json.Unmarshal(b, &m)

	for key, obj := range m {
		err = rc.PutJsonObj(key, obj)
		if err != nil {
			return err
		}
	}

	return nil
}
