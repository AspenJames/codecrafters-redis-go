package cache

import "time"

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

func (c *Cache) Set(key string, value string, expiry time.Time) {
	c.cache[key] = &val{val: value, exp: expiry}
}
