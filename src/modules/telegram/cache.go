package telegram

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"sync"
	"time"
)

type cacheEntry struct {
	musicURL string
	title    string
	expires  time.Time
}

// Cache maps short IDs to music URLs with TTL expiry.
type Cache struct {
	mu   sync.Mutex
	data map[string]cacheEntry
	ttl  time.Duration
	quit chan struct{}
}

func NewCache() *Cache {
	return NewCacheWithTTL(time.Hour)
}

// NewCacheWithTTL is exposed for tests that need a shorter TTL.
func NewCacheWithTTL(ttl time.Duration) *Cache {
	c := &Cache{data: make(map[string]cacheEntry), ttl: ttl, quit: make(chan struct{})}
	go c.janitor()
	return c
}

// Close stops the background janitor goroutine.
func (c *Cache) Close() {
	close(c.quit)
}

func (c *Cache) Put(musicURL, title string) string {
	id := randomID()
	c.mu.Lock()
	c.data[id] = cacheEntry{musicURL: musicURL, title: title, expires: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return id
}

func (c *Cache) Get(id string) (musicURL, title string, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, exists := c.data[id]
	if !exists || time.Now().After(e.expires) {
		return "", "", false
	}
	return e.musicURL, e.title, true
}

func (c *Cache) janitor() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-c.quit:
			return
		case now := <-ticker.C:
			c.mu.Lock()
			for k, v := range c.data {
				if now.After(v.expires) {
					delete(c.data, k)
				}
			}
			c.mu.Unlock()
		}
	}
}

func randomID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		log.Panicf("crypto/rand unavailable: %v", err)
	}
	return hex.EncodeToString(b)
}
