package wbapicredentials

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// Cache represents an in memory cache for keys caching
type Cache struct {
	sync.RWMutex
	cache *cache.Cache
}

// NewCache creates a cache for caching just keys
func NewCache(defaultExpiration, cleanUpInterval time.Duration) *Cache {
	return &Cache{cache: cache.New(defaultExpiration, cleanUpInterval)}
}

// Contains return true if cache contains item
func (c *Cache) Contains(credentialsHash string) bool {
	c.RLock()
	defer c.RUnlock()
	_, found := c.cache.Get(credentialsHash)
	return found
}

// Put stores crenedtials in cache
func (c *Cache) Put(credentialsHash string) {
	c.Lock()
	defer c.Unlock()
	c.cache.Set(credentialsHash, nil, cache.DefaultExpiration)
}

// Flush clears the cache
func (c *Cache) Flush() {
	c.Lock()
	defer c.Unlock()
	c.cache.Flush()
}
