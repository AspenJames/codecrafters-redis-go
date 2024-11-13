package cache

import (
	"fmt"
	"regexp"
	"time"
)

type Cache struct {
	cache map[string]*val
}

var defaultCache *Cache = &Cache{cache: map[string]*val{}}

func GetDefaultCache() *Cache {
	return defaultCache
}

type val struct {
	val string
	exp time.Time
}

func (v *val) isExpired() bool {
	return !v.exp.IsZero() && time.Now().After(v.exp)
}

func (c *Cache) KeyExists(key string) bool {
	val, exists := c.cache[key]
	return exists && !val.isExpired()
}

func (c *Cache) Get(key string) (string, bool) {
	val, ok := c.cache[key]
	if !ok {
		return "", false
	}
	if val.isExpired() {
		delete(c.cache, key)
		return "", false
	}
	return val.val, ok
}

func (c *Cache) GetKeys(pattern *regexp.Regexp) []string {
	var keys []string
	for k := range c.cache {
		if pattern.Match([]byte(k)) {
			keys = append(keys, k)
		}
	}
	return keys
}

func (c *Cache) Set(key string, value string, expiry time.Time) {
	c.cache[key] = &val{val: value, exp: expiry}
}

// elems is a slice of slices. Each slice is a k/v pair with an optional expiry -- k, v[, e]
func (c *Cache) LoadRDB(resp interface{}) error {
	if resp == nil {
		return nil
	}

	elems, ok := resp.([][]interface{})
	if !ok {
		return fmt.Errorf("unexpected RDBParser response format")
	}
	// Dump current cache
	c.cache = make(map[string]*val)

	for _, elem := range elems {
		var (
			key, val []byte
			expiry   time.Time
			ok       bool
		)
		switch len(elem) {
		case 3: // k/v with expiry
			expiry, ok = elem[2].(time.Time)
			if !ok {
				return fmt.Errorf("wrong time format: %#v", elem[2])
			}
			fallthrough
		case 2: // k/v
			key, ok = elem[0].([]byte)
			if !ok {
				return fmt.Errorf("improper key type: %#v", elem[0])
			}
			val, ok = elem[1].([]byte)
			if !ok {
				return fmt.Errorf("improper val type: %#v", elem[0])
			}
		default: // unexpected length
			return fmt.Errorf("unexpected k/v pair length %#v", len(elem))
		}
		// Load into cache.
		c.Set(string(key), string(val), expiry)
	}
	return nil
}
