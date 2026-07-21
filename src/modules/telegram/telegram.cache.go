package telegram

import (
	"crypto/rand"
	"encoding/hex"
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
}

// NewCache returns a Cache with a 1-hour TTL.
func NewCache() *Cache {
	return NewCacheWithTTL(time.Hour)
}

// NewCacheWithTTL returns a Cache with a custom TTL (used in tests).
func NewCacheWithTTL(ttl time.Duration) *Cache {
	c := &Cache{data: make(map[string]cacheEntry), ttl: ttl}
	go c.janitor()
	return c
}

// Put stores a music URL and title, returning an 8-hex-char ID.
func (c *Cache) Put(musicURL, title string) string {
	id := randomID()
	c.mu.Lock()
	c.data[id] = cacheEntry{musicURL: musicURL, title: title, expires: time.Now().Add(c.ttl)}
	c.mu.Unlock()
	return id
}

// Get retrieves a music URL by ID. Returns ok=false if expired or not found.
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
	for range ticker.C {
		now := time.Now()
		c.mu.Lock()
		for k, v := range c.data {
			if now.After(v.expires) {
				delete(c.data, k)
			}
		}
		c.mu.Unlock()
	}
}

func randomID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
