package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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

func (c *LocalCache) Get(key string) (interface{}, bool, error) {
	c.mu.RLock()
	data, ok := c.data[key]
	c.mu.RUnlock()
	return data, ok, nil
}

func (c *LocalCache) GetAll(keys []string) map[string]interface{} {
	entries := make(map[string]interface{})
	for _, k := range keys {
		entries[k], _, _ = c.Get(k)
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

func (c *LocalCache) GetKeys(prefix string) ([]string, error) {
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		if strings.Contains(k, prefix) {
			keys = append(keys, k)
		}
	}

	return keys, nil
}

func (c *LocalCache) GetJsonObj(key string) ([]byte, error) {
	jf, err := os.Open("./local.json")
	if err != nil {
		return nil, err
	}
	defer jf.Close()
	b, _ := ioutil.ReadAll(jf)

	return b, err
}

func (c *LocalCache) GetJsonObjKeys(objKey string) (interface{}, error) {
	jf, err := os.Open("./local.json")
	if err != nil {
		return nil, err
	}
	defer jf.Close()
	b, _ := ioutil.ReadAll(jf)

	var m map[string]interface{}
	json.Unmarshal(b, &m)

	var keys []string
	if obj, ok := m[objKey]; ok {
		for key := range obj.(map[string]interface{}) {
			keys = append(keys, key)
		}

		return keys, nil
	}

	return nil, fmt.Errorf("JSON object not found")
}

func (c *LocalCache) PutJsonObj(key string, object interface{}) error {
	jf, err := os.Open("./local.json")
	if err != nil {
		return err
	}
	defer jf.Close()
	b, _ := ioutil.ReadAll(jf)

	var m map[string]interface{}
	json.Unmarshal(b, &m)

	m[key] = object
	jm, _ := json.MarshalIndent(m, "", "    ")

	jf.Truncate(0)

	return ioutil.WriteFile("./local.json", jm, 0644)
}

func (c *LocalCache) PutJsonFromLocalFile() error {
	return fmt.Errorf("you're using local cache, silly")
}
