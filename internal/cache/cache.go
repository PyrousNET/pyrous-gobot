package cache

import (
	"net/url"
)

type Cache interface {
	Put(key string, value interface{})
	PutAll(map[string]interface{})
	Get(key string) (interface{}, bool, error)
	GetAll(keys []string) map[string]interface{}
	Clean(key string)
	GetKeys(prefix string) ([]string, error)
	CleanAll()
	GetJsonObj(key string) ([]byte, error)
	GetJsonObjKeys(key string) (interface{}, error)
	PutJsonObj(key string, object interface{}) error
	PutJsonFromLocalFile() error
}

func GetCachingMechanism(connStr string) Cache {
	uri, _ := url.Parse(connStr)

	switch uri.Scheme {
	case "redis":
		return GetRedisCache(connStr)
	default:
		return GetLocalCache()
	}
}
