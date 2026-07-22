package telegram_test

import (
	"testing"
	"time"

	"telegram-bot/src/modules/telegram"
)

func TestCache_PutAndGet(t *testing.T) {
	c := telegram.NewCache()
	id := c.Put("https://example.com/music.mp3", "Test Song")
	if id == "" {
		t.Fatal("expected non-empty ID")
	}
	musicURL, title, ok := c.Get(id)
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if musicURL != "https://example.com/music.mp3" {
		t.Errorf("musicURL: want https://example.com/music.mp3, got %s", musicURL)
	}
	if title != "Test Song" {
		t.Errorf("title: want Test Song, got %s", title)
	}
}

func TestCache_Get_NotFound(t *testing.T) {
	c := telegram.NewCache()
	_, _, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for nonexistent ID")
	}
}

func TestCache_Get_Expired(t *testing.T) {
	c := telegram.NewCacheWithTTL(50 * time.Millisecond)
	id := c.Put("https://example.com/music.mp3", "Test Song")
	time.Sleep(100 * time.Millisecond)
	_, _, ok := c.Get(id)
	if ok {
		t.Error("expected ok=false for expired entry")
	}
}

func TestCache_IDsAreUnique(t *testing.T) {
	c := telegram.NewCache()
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := c.Put("https://example.com/music.mp3", "Song")
		if seen[id] {
			t.Fatalf("duplicate ID generated: %s", id)
		}
		seen[id] = true
	}
}
