package torrent

import (
	"sync"
	"time"
)

type CacheItem struct {
	Results []Result
	Created time.Time
}

type Cache interface {
	Set(key string, results []Result) error
	Get(key string) ([]Result, time.Time, bool, error)
}

var _ Cache = (*InMemCache)(nil)

type InMemCache struct {
	cache map[string]CacheItem
	*sync.RWMutex
}

func NewInMemCache() *InMemCache {
	return &InMemCache{
		map[string]CacheItem{}, &sync.RWMutex{},
	}
}

func (c *InMemCache) Set(key string, results []Result) error {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	c.cache[key] = CacheItem{
		Results: results,
		Created: time.Now(),
	}
	return nil
}

func (c *InMemCache) Get(key string) ([]Result, time.Time, bool, error) {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	cacheItem, found := c.cache[key]
	return cacheItem.Results, cacheItem.Created, found, nil
}
