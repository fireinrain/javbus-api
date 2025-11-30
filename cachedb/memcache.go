package cachedb

import (
	"sync"
	"time"
)

// Cache 内存缓存实现
type Cache struct {
	mu           sync.RWMutex
	items        map[string]*cacheItem
	defaultTTL   time.Duration
	cleanupTimer *time.Ticker
}

// cacheItem 缓存项
type cacheItem struct {
	value      interface{}
	expiration int64
}

// NewCache 创建新的缓存实例
func NewCache(defaultTTL time.Duration, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		items:      make(map[string]*cacheItem),
		defaultTTL: defaultTTL,
	}

	if cleanupInterval > 0 {
		c.cleanupTimer = time.NewTicker(cleanupInterval)
		go c.cleanupLoop()
	}

	return c
}

// Set 设置缓存项
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	} else if c.defaultTTL > 0 {
		expiration = time.Now().Add(c.defaultTTL).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = &cacheItem{
		value:      value,
		expiration: expiration,
	}
}

// Get 获取缓存项
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		// 过期但不立即删除，由清理协程处理
		return nil, false
	}

	return item.value, true
}

// cleanupLoop 定期清理过期缓存
func (c *Cache) cleanupLoop() {
	for range c.cleanupTimer.C {
		c.cleanup()
	}
}

// cleanup 清理过期缓存
func (c *Cache) cleanup() {
	now := time.Now().UnixNano()

	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.items {
		if v.expiration > 0 && now > v.expiration {
			delete(c.items, k)
		}
	}
}
