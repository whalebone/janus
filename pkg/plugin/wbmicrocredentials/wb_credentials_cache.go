package wbmicrocredentials

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

// CachedCredentials struct for caching credentials
type CachedCredentials struct {
	ClientID     string
	LoginSuccess bool
}

// CredentialsCache represents a in memory cache for credentials
type CredentialsCache struct {
	sync.RWMutex
	credentailsCache *cache.Cache
}

// NewCredentialsCache creates a cache for credentials
func NewCredentialsCache(defaultExpiration, cleanUpInterval time.Duration) *CredentialsCache {
	return &CredentialsCache{credentailsCache: cache.New(defaultExpiration, cleanUpInterval)}
}

// Get fetches credentials from cache
func (c *CredentialsCache) Get(credentialsHash string) (*CachedCredentials, bool) {
	c.RLock()
	defer c.RUnlock()
	if item, found := c.credentailsCache.Get(credentialsHash); found {
		return item.(*CachedCredentials), true
	}
	return nil, false
}

// Put fetches credentials from cache
func (c *CredentialsCache) Put(credentialsHash string, credentials *CachedCredentials) {
	c.Lock()
	defer c.Unlock()
	c.credentailsCache.Set(credentialsHash, credentials, cache.DefaultExpiration)
}

// Flush clears the cache
func (c *CredentialsCache) Flush() {
	c.Lock()
	defer c.Unlock()
	c.credentailsCache.Flush()
}
