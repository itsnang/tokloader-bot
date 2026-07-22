# Telegram TikTok Bot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Telegram bot that resolves TikTok URLs via tikwm.com and sends back a watermark-free video or photo slideshow with an inline MP3 download button.

**Architecture:** Two Go modules under `src/modules/` — `tiktok` handles tikwm.com API calls, `telegram` handles the bot loop and message/callback routing. They share state through a `tiktok.Service` interface injected via a manually-written wire initializer. An in-memory TTL cache maps 8-hex-char IDs to music URLs to work around Telegram's 64-byte callback data limit.

**Tech Stack:** Go 1.22+, `github.com/go-telegram-bot-api/telegram-bot-api/v5`, `github.com/google/wire`, `github.com/joho/godotenv`

---

## File map

| File | Responsibility |
|------|---------------|
| `main.go` | Load `.env`, build bot via `initBot`, run until signal |
| `wire.go` | Wire injector declaration (build-tag guarded) |
| `wire_gen.go` | Manual wire output — no CLI needed |
| `src/modules/tiktok/tiktok.dto.go` | `InfoResponse` struct + `IsImage()` |
| `src/modules/tiktok/tiktok.client.go` | HTTP client for tikwm.com API |
| `src/modules/tiktok/tiktok.service.go` | `Service` interface + impl with validation |
| `src/modules/tiktok/tiktok.provider.go` | Wire `ProviderSet` |
| `src/modules/telegram/telegram.cache.go` | Short-ID → music URL TTL cache |
| `src/modules/telegram/telegram.handler.go` | `OnMessage` + `OnCallback` + `Sender` interface |
| `src/modules/telegram/telegram.bot.go` | `BotAPI` setup + long-polling update loop |
| `src/modules/telegram/telegram.provider.go` | Wire `ProviderSet` |

---

### Task 1: Initialize Go module and project skeleton

**Files:**
- Create: `go.mod`
- Create: `.env.example`
- Create: `.gitignore`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/samnanglorn/Desktop/telegram-bot
go mod init telegram-bot
```

Expected: `go.mod` created with `module telegram-bot`.

- [ ] **Step 2: Add dependencies**

```bash
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
go get github.com/google/wire
go get github.com/joho/godotenv
```

- [ ] **Step 3: Create directory structure**

```bash
mkdir -p src/modules/tiktok src/modules/telegram
```

- [ ] **Step 4: Create `.env.example`**

```
TELEGRAM_BOT_TOKEN=your_bot_token_here
TIKWM_BASE_URL=https://www.tikwm.com
```

- [ ] **Step 5: Create `.gitignore`**

```
.env
```

- [ ] **Step 6: Commit**

```bash
git init
git add go.mod go.sum .env.example .gitignore
git commit -m "chore: initialize go module with dependencies"
```

---

### Task 2: TikTok DTO

**Files:**
- Create: `src/modules/tiktok/tiktok.dto.go`
- Create: `src/modules/tiktok/tiktok_dto_test.go`

- [ ] **Step 1: Write the failing test**

`src/modules/tiktok/tiktok_dto_test.go`:
```go
package tiktok_test

import (
	"testing"

	"telegram-bot/src/modules/tiktok"
)

func TestIsImage_WithImages(t *testing.T) {
	r := &tiktok.InfoResponse{Images: []string{"https://example.com/img1.jpg"}}
	if !r.IsImage() {
		t.Error("expected IsImage() = true when Images is non-empty")
	}
}

func TestIsImage_WithoutImages(t *testing.T) {
	r := &tiktok.InfoResponse{}
	if r.IsImage() {
		t.Error("expected IsImage() = false when Images is empty")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./src/modules/tiktok/... -run TestIsImage -v
```

Expected: FAIL — package not found or type undefined.

- [ ] **Step 3: Write the DTO**

`src/modules/tiktok/tiktok.dto.go`:
```go
package tiktok

// InfoResponse holds resolved TikTok post data.
type InfoResponse struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Duration    int      `json:"duration"`
	Cover       string   `json:"cover"`
	NoWatermark string   `json:"no_watermark"`
	Watermark   string   `json:"watermark"`
	Music       string   `json:"music"`
	Images      []string `json:"images"`
}

// IsImage reports whether the post is a photo slideshow rather than a video.
func (r *InfoResponse) IsImage() bool {
	return len(r.Images) > 0
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./src/modules/tiktok/... -run TestIsImage -v
```

Expected: PASS (both tests).

- [ ] **Step 5: Commit**

```bash
git add src/modules/tiktok/tiktok.dto.go src/modules/tiktok/tiktok_dto_test.go
git commit -m "feat(tiktok): add InfoResponse DTO with IsImage helper"
```

---

### Task 3: TikTok HTTP Client

**Files:**
- Create: `src/modules/tiktok/tiktok.client.go`
- Create: `src/modules/tiktok/tiktok_client_test.go`

- [ ] **Step 1: Write the failing test**

`src/modules/tiktok/tiktok_client_test.go`:
```go
package tiktok_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"telegram-bot/src/modules/tiktok"
)

func newTestServer(t *testing.T, response map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

func TestClient_Info_Video(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"code": 0,
		"msg":  "success",
		"data": map[string]any{
			"id":       "123",
			"title":    "Test video",
			"author":   map[string]any{"unique_id": "testuser"},
			"duration": 30,
			"cover":    "https://example.com/cover.jpg",
			"hdplay":   "https://example.com/video_hd.mp4",
			"play":     "https://example.com/video.mp4",
			"wmplay":   "https://example.com/video_wm.mp4",
			"music":    "https://example.com/music.mp3",
		},
	})
	defer srv.Close()

	client := tiktok.NewClient(srv.URL)
	info, err := client.Info(context.Background(), "https://tiktok.com/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ID != "123" {
		t.Errorf("ID: want 123, got %s", info.ID)
	}
	if info.NoWatermark != "https://example.com/video_hd.mp4" {
		t.Errorf("NoWatermark: want hdplay URL, got %s", info.NoWatermark)
	}
	if info.Music != "https://example.com/music.mp3" {
		t.Errorf("Music: want music URL, got %s", info.Music)
	}
	if info.IsImage() {
		t.Error("expected IsImage() = false for video post")
	}
}

func TestClient_Info_HDPlayFallback(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"code": 0,
		"msg":  "success",
		"data": map[string]any{
			"id":     "124",
			"title":  "No HD video",
			"author": map[string]any{"unique_id": "testuser"},
			"hdplay": "",
			"play":   "https://example.com/video.mp4",
			"wmplay": "https://example.com/video_wm.mp4",
			"music":  "https://example.com/music.mp3",
		},
	})
	defer srv.Close()

	client := tiktok.NewClient(srv.URL)
	info, err := client.Info(context.Background(), "https://tiktok.com/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.NoWatermark != "https://example.com/video.mp4" {
		t.Errorf("NoWatermark: want play fallback URL, got %s", info.NoWatermark)
	}
}

func TestClient_Info_Slideshow(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"code": 0,
		"msg":  "success",
		"data": map[string]any{
			"id":     "456",
			"title":  "Test slideshow",
			"author": map[string]any{"unique_id": "testuser"},
			"images": []string{
				"https://example.com/img1.jpg",
				"https://example.com/img2.jpg",
			},
			"music": "https://example.com/music.mp3",
		},
	})
	defer srv.Close()

	client := tiktok.NewClient(srv.URL)
	info, err := client.Info(context.Background(), "https://tiktok.com/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.IsImage() {
		t.Error("expected IsImage() = true for slideshow post")
	}
	if len(info.Images) != 2 {
		t.Errorf("Images: want 2, got %d", len(info.Images))
	}
}

func TestClient_Info_APIError(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"code": -1,
		"msg":  "video not found",
	})
	defer srv.Close()

	client := tiktok.NewClient(srv.URL)
	_, err := client.Info(context.Background(), "https://tiktok.com/test")
	if err == nil {
		t.Error("expected error for non-zero API code")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./src/modules/tiktok/... -run TestClient -v
```

Expected: FAIL — `NewClient` undefined.

- [ ] **Step 3: Write the client**

`src/modules/tiktok/tiktok.client.go`:
```go
package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Client calls the tikwm.com resolver API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a Client. baseURL should be "https://www.tikwm.com" in production.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

type resolverAuthor struct {
	UniqueID string `json:"unique_id"`
}

type resolverData struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Author   resolverAuthor `json:"author"`
	Duration int            `json:"duration"`
	Cover    string         `json:"cover"`
	HDPlay   string         `json:"hdplay"`
	Play     string         `json:"play"`
	WMPlay   string         `json:"wmplay"`
	Music    string         `json:"music"`
	Images   []string       `json:"images"`
}

type resolverResponse struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data resolverData `json:"data"`
}

// Info resolves a TikTok URL and returns post metadata.
func (c *Client) Info(ctx context.Context, tikTokURL string) (*InfoResponse, error) {
	form := url.Values{}
	form.Set("url", tikTokURL)
	form.Set("hd", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tikwm request: %w", err)
	}
	defer resp.Body.Close()

	var raw resolverResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if raw.Code != 0 {
		return nil, fmt.Errorf("tikwm error: %s", raw.Msg)
	}

	noWatermark := raw.Data.HDPlay
	if noWatermark == "" {
		noWatermark = raw.Data.Play
	}

	return &InfoResponse{
		ID:          raw.Data.ID,
		Title:       raw.Data.Title,
		Author:      raw.Data.Author.UniqueID,
		Duration:    raw.Data.Duration,
		Cover:       raw.Data.Cover,
		NoWatermark: noWatermark,
		Watermark:   raw.Data.WMPlay,
		Music:       raw.Data.Music,
		Images:      raw.Data.Images,
	}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./src/modules/tiktok/... -run TestClient -v
```

Expected: PASS (all 4 tests).

- [ ] **Step 5: Commit**

```bash
git add src/modules/tiktok/tiktok.client.go src/modules/tiktok/tiktok_client_test.go
git commit -m "feat(tiktok): add tikwm HTTP client"
```

---

### Task 4: TikTok Service

**Files:**
- Create: `src/modules/tiktok/tiktok.service.go`
- Create: `src/modules/tiktok/tiktok_service_test.go`

- [ ] **Step 1: Write the failing test**

`src/modules/tiktok/tiktok_service_test.go`:
```go
package tiktok_test

import (
	"context"
	"errors"
	"testing"

	"telegram-bot/src/modules/tiktok"
)

type mockClient struct {
	info *tiktok.InfoResponse
	err  error
}

func (m *mockClient) Info(_ context.Context, _ string) (*tiktok.InfoResponse, error) {
	return m.info, m.err
}

func TestService_Info_Success(t *testing.T) {
	mc := &mockClient{info: &tiktok.InfoResponse{ID: "1", NoWatermark: "https://cdn.example.com/v.mp4"}}
	svc := tiktok.NewService(mc)
	info, err := svc.Info(context.Background(), "https://tiktok.com/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ID != "1" {
		t.Errorf("ID: want 1, got %s", info.ID)
	}
}

func TestService_Info_ClientError(t *testing.T) {
	mc := &mockClient{err: errors.New("network error")}
	svc := tiktok.NewService(mc)
	_, err := svc.Info(context.Background(), "https://tiktok.com/test")
	if err == nil {
		t.Error("expected error from client")
	}
}

func TestService_Info_EmptyVideoURL(t *testing.T) {
	mc := &mockClient{info: &tiktok.InfoResponse{ID: "2", NoWatermark: ""}}
	svc := tiktok.NewService(mc)
	_, err := svc.Info(context.Background(), "https://tiktok.com/test")
	if err == nil {
		t.Error("expected error when NoWatermark is empty and Images is empty")
	}
}

func TestService_Info_SlideshowNoWatermarkOK(t *testing.T) {
	mc := &mockClient{info: &tiktok.InfoResponse{
		ID:     "3",
		Images: []string{"https://example.com/img.jpg"},
		Music:  "https://example.com/music.mp3",
	}}
	svc := tiktok.NewService(mc)
	info, err := svc.Info(context.Background(), "https://tiktok.com/test")
	if err != nil {
		t.Fatalf("unexpected error for slideshow: %v", err)
	}
	if !info.IsImage() {
		t.Error("expected IsImage() = true")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./src/modules/tiktok/... -run TestService -v
```

Expected: FAIL — `NewService` undefined.

- [ ] **Step 3: Write the service**

`src/modules/tiktok/tiktok.service.go`:
```go
package tiktok

import (
	"context"
	"fmt"
)

// InfoClient is the interface the service depends on (satisfied by *Client).
type InfoClient interface {
	Info(ctx context.Context, url string) (*InfoResponse, error)
}

// Service resolves TikTok post metadata.
type Service interface {
	Info(ctx context.Context, url string) (*InfoResponse, error)
}

type service struct {
	client InfoClient
}

// NewService wraps client with validation logic.
func NewService(client InfoClient) Service {
	return &service{client: client}
}

func (s *service) Info(ctx context.Context, url string) (*InfoResponse, error) {
	info, err := s.client.Info(ctx, url)
	if err != nil {
		return nil, err
	}
	if info.NoWatermark == "" && !info.IsImage() {
		return nil, fmt.Errorf("resolver returned no usable video URL")
	}
	return info, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./src/modules/tiktok/... -run TestService -v
```

Expected: PASS (all 4 tests).

- [ ] **Step 5: Commit**

```bash
git add src/modules/tiktok/tiktok.service.go src/modules/tiktok/tiktok_service_test.go
git commit -m "feat(tiktok): add service interface with validation"
```

---

### Task 5: TikTok Wire Provider

**Files:**
- Create: `src/modules/tiktok/tiktok.provider.go`

- [ ] **Step 1: Write the provider**

`src/modules/tiktok/tiktok.provider.go`:
```go
package tiktok

import "github.com/google/wire"

// ProviderSet binds tiktok constructors for Wire.
var ProviderSet = wire.NewSet(
	NewClient,
	wire.Bind(new(InfoClient), new(*Client)),
	NewService,
)
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./src/modules/tiktok/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add src/modules/tiktok/tiktok.provider.go
git commit -m "feat(tiktok): add wire provider set"
```

---

### Task 6: Telegram Cache

**Files:**
- Create: `src/modules/telegram/telegram.cache.go`
- Create: `src/modules/telegram/telegram_cache_test.go`

- [ ] **Step 1: Write the failing test**

`src/modules/telegram/telegram_cache_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./src/modules/telegram/... -run TestCache -v
```

Expected: FAIL — `NewCache` undefined.

- [ ] **Step 3: Write the cache**

`src/modules/telegram/telegram.cache.go`:
```go
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
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./src/modules/telegram/... -run TestCache -v
```

Expected: PASS (all 4 tests).

- [ ] **Step 5: Commit**

```bash
git add src/modules/telegram/telegram.cache.go src/modules/telegram/telegram_cache_test.go
git commit -m "feat(telegram): add TTL cache for music URL short IDs"
```

---

### Task 7: Telegram Handler

**Files:**
- Create: `src/modules/telegram/telegram.handler.go`
- Create: `src/modules/telegram/telegram_handler_test.go`

- [ ] **Step 1: Write the failing test**

`src/modules/telegram/telegram_handler_test.go`:
```go
package telegram_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/src/modules/telegram"
	"telegram-bot/src/modules/tiktok"
)

// mockSender captures sent messages for assertions.
// Uses interface{} because MediaGroupConfig is not Chattable.
type mockSender struct {
	mu       sync.Mutex
	sent     []interface{}
	requests []tgbotapi.Chattable
}

func (m *mockSender) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	m.mu.Lock()
	m.sent = append(m.sent, c)
	m.mu.Unlock()
	return tgbotapi.Message{MessageID: 1}, nil
}

func (m *mockSender) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	m.mu.Lock()
	m.requests = append(m.requests, c)
	m.mu.Unlock()
	return &tgbotapi.APIResponse{Ok: true}, nil
}

func (m *mockSender) SendMediaGroup(config tgbotapi.MediaGroupConfig) ([]tgbotapi.Message, error) {
	m.mu.Lock()
	m.sent = append(m.sent, config)
	m.mu.Unlock()
	return []tgbotapi.Message{}, nil
}

// mockService returns preset results.
type mockService struct {
	info *tiktok.InfoResponse
	err  error
}

func (m *mockService) Info(_ context.Context, _ string) (*tiktok.InfoResponse, error) {
	return m.info, m.err
}

func newMsg(chatID int64, text string) *tgbotapi.Message {
	return &tgbotapi.Message{
		Chat: &tgbotapi.Chat{ID: chatID},
		Text: text,
	}
}

func TestHandler_OnMessage_Start(t *testing.T) {
	sender := &mockSender{}
	h := telegram.NewHandler(sender, &mockService{}, telegram.NewCache())

	h.OnMessage(context.Background(), newMsg(42, "/start"))

	if len(sender.sent) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(sender.sent))
	}
}

func TestHandler_OnMessage_NonTikTok(t *testing.T) {
	sender := &mockSender{}
	h := telegram.NewHandler(sender, &mockService{}, telegram.NewCache())

	h.OnMessage(context.Background(), newMsg(42, "https://youtube.com/watch?v=xxx"))

	if len(sender.sent) != 1 {
		t.Fatalf("expected 1 message (hint), got %d", len(sender.sent))
	}
}

func TestHandler_OnMessage_Video(t *testing.T) {
	sender := &mockSender{}
	svc := &mockService{info: &tiktok.InfoResponse{
		ID:          "1",
		Title:       "Test video",
		Author:      "testuser",
		NoWatermark: "https://cdn.example.com/video.mp4",
		Music:       "https://cdn.example.com/music.mp3",
	}}
	h := telegram.NewHandler(sender, svc, telegram.NewCache())

	h.OnMessage(context.Background(), newMsg(42, "https://www.tiktok.com/@user/video/1"))

	// sent[0] = "Processing...", sent[1] = VideoConfig
	found := false
	for _, s := range sender.sent {
		if _, ok := s.(tgbotapi.VideoConfig); ok {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a VideoConfig to be sent; got %d messages: %T", len(sender.sent), sender.sent)
	}
}

func TestHandler_OnMessage_Slideshow(t *testing.T) {
	sender := &mockSender{}
	svc := &mockService{info: &tiktok.InfoResponse{
		ID:     "2",
		Title:  "Slideshow",
		Author: "testuser",
		Images: []string{"https://cdn.example.com/img1.jpg", "https://cdn.example.com/img2.jpg"},
		Music:  "https://cdn.example.com/music.mp3",
	}}
	h := telegram.NewHandler(sender, svc, telegram.NewCache())

	h.OnMessage(context.Background(), newMsg(42, "https://www.tiktok.com/@user/photo/2"))

	found := false
	for _, s := range sender.sent {
		if _, ok := s.(tgbotapi.MediaGroupConfig); ok {
			found = true
		}
	}
	if !found {
		t.Error("expected a MediaGroupConfig to be sent for slideshow")
	}
}

func TestHandler_OnMessage_ServiceError(t *testing.T) {
	sender := &mockSender{}
	svc := &mockService{err: errors.New("not found")}
	h := telegram.NewHandler(sender, svc, telegram.NewCache())

	h.OnMessage(context.Background(), newMsg(42, "https://www.tiktok.com/@user/video/1"))

	// sent[0] = "Processing...", sent[1] = error reply
	if len(sender.sent) < 2 {
		t.Fatalf("expected at least 2 sends (processing + error), got %d", len(sender.sent))
	}
}

func TestHandler_OnCallback_MP3(t *testing.T) {
	sender := &mockSender{}
	cache := telegram.NewCache()
	id := cache.Put("https://cdn.example.com/music.mp3", "Test Song")
	h := telegram.NewHandler(sender, &mockService{}, cache)

	cb := &tgbotapi.CallbackQuery{
		ID:   "cb1",
		Data: "mp3:" + id,
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 42},
		},
	}
	h.OnCallback(context.Background(), cb)

	found := false
	for _, s := range sender.sent {
		if _, ok := s.(tgbotapi.AudioConfig); ok {
			found = true
		}
	}
	if !found {
		t.Error("expected an AudioConfig to be sent")
	}
}

func TestHandler_OnCallback_ExpiredID(t *testing.T) {
	sender := &mockSender{}
	h := telegram.NewHandler(sender, &mockService{}, telegram.NewCache())

	cb := &tgbotapi.CallbackQuery{
		ID:   "cb2",
		Data: "mp3:deadbeef",
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: 42},
		},
	}
	h.OnCallback(context.Background(), cb)

	// sent[0] = expired notice (callback answer goes to requests)
	if len(sender.sent) != 1 {
		t.Fatalf("expected 1 message (expired notice), got %d", len(sender.sent))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./src/modules/telegram/... -run TestHandler -v
```

Expected: FAIL — `NewHandler` undefined.

- [ ] **Step 3: Write the handler**

`src/modules/telegram/telegram.handler.go`:
```go
package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/src/modules/tiktok"
)

// Sender is the subset of tgbotapi.BotAPI methods used by the handler.
// Defined as an interface so the handler can be tested without a live bot.
type Sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	SendMediaGroup(config tgbotapi.MediaGroupConfig) ([]tgbotapi.Message, error)
}

// Handler processes Telegram messages and callback queries.
type Handler struct {
	sender  Sender
	service tiktok.Service
	cache   *Cache
}

// NewHandler creates a Handler.
func NewHandler(sender Sender, service tiktok.Service, cache *Cache) *Handler {
	return &Handler{sender: sender, service: service, cache: cache}
}

// OnMessage handles an incoming text message.
func (h *Handler) OnMessage(ctx context.Context, msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.Text)

	if text == "/start" {
		h.send(tgbotapi.NewMessage(msg.Chat.ID, "Send me a TikTok link and I'll grab it for you 🎬"))
		return
	}
	if !strings.Contains(strings.ToLower(text), "tiktok.com") {
		h.send(tgbotapi.NewMessage(msg.Chat.ID, "That doesn't look like a TikTok link 🤔"))
		return
	}

	wait, _ := h.sender.Send(tgbotapi.NewMessage(msg.Chat.ID, "Processing... ⏳"))
	defer h.sender.Request(tgbotapi.NewDeleteMessage(msg.Chat.ID, wait.MessageID))

	info, err := h.service.Info(ctx, text)
	if err != nil {
		h.send(tgbotapi.NewMessage(msg.Chat.ID, "Couldn't fetch that video 😕 "+err.Error()))
		return
	}

	mp3Btn := h.mp3Button(info)
	caption := buildCaption(info)

	if info.IsImage() {
		h.sendImages(msg.Chat.ID, info.Images)
		out := tgbotapi.NewMessage(msg.Chat.ID, caption)
		if mp3Btn != nil {
			out.ReplyMarkup = mp3Btn
		}
		h.send(out)
	} else {
		video := tgbotapi.NewVideo(msg.Chat.ID, tgbotapi.FileURL(info.NoWatermark))
		video.Caption = caption
		if mp3Btn != nil {
			video.ReplyMarkup = mp3Btn
		}
		h.send(video)
	}
}

// OnCallback handles the MP3 button tap.
func (h *Handler) OnCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	defer h.sender.Request(tgbotapi.NewCallback(cb.ID, ""))

	if !strings.HasPrefix(cb.Data, "mp3:") {
		return
	}
	id := strings.TrimPrefix(cb.Data, "mp3:")

	musicURL, title, ok := h.cache.Get(id)
	if !ok || musicURL == "" {
		h.send(tgbotapi.NewMessage(cb.Message.Chat.ID, "That MP3 link expired ⌛ send the video again."))
		return
	}

	audio := tgbotapi.NewAudio(cb.Message.Chat.ID, tgbotapi.FileURL(musicURL))
	if title != "" {
		audio.Title = title
	}
	h.send(audio)
}

func (h *Handler) mp3Button(info *tiktok.InfoResponse) *tgbotapi.InlineKeyboardMarkup {
	if info.Music == "" {
		return nil
	}
	id := h.cache.Put(info.Music, info.Title)
	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎵 Download MP3", "mp3:"+id),
		),
	)
	return &markup
}

func (h *Handler) sendImages(chatID int64, images []string) {
	media := make([]interface{}, 0, len(images))
	for _, img := range images {
		media = append(media, tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(img)))
	}
	for i := 0; i < len(media); i += 10 {
		end := i + 10
		if end > len(media) {
			end = len(media)
		}
		h.sender.SendMediaGroup(tgbotapi.NewMediaGroup(chatID, media[i:end]))
	}
}

func (h *Handler) send(c tgbotapi.Chattable) {
	h.sender.Send(c) //nolint:errcheck
}

func buildCaption(info *tiktok.InfoResponse) string {
	if info.Author != "" {
		return fmt.Sprintf("%s\n\n👤 %s", info.Title, info.Author)
	}
	return info.Title
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./src/modules/telegram/... -run TestHandler -v
```

Expected: PASS (all 7 tests).

- [ ] **Step 5: Commit**

```bash
git add src/modules/telegram/telegram.handler.go src/modules/telegram/telegram_handler_test.go
git commit -m "feat(telegram): add message and callback handler"
```

---

### Task 8: Telegram Bot

**Files:**
- Create: `src/modules/telegram/telegram.bot.go`

- [ ] **Step 1: Write the bot**

`src/modules/telegram/telegram.bot.go`:
```go
package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/src/modules/tiktok"
)

// Bot wraps the Telegram BotAPI and runs the update loop.
type Bot struct {
	api     *tgbotapi.BotAPI
	handler *Handler
}

// NewBot creates a Bot from the given token, service, and cache.
func NewBot(token string, service tiktok.Service, cache *Cache) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	handler := NewHandler(api, service, cache)
	return &Bot{api: api, handler: handler}, nil
}

// Start runs the long-polling update loop. Blocks until ctx is cancelled.
func (b *Bot) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			switch {
			case update.CallbackQuery != nil:
				go b.handler.OnCallback(ctx, update.CallbackQuery)
			case update.Message != nil && update.Message.Text != "":
				go b.handler.OnMessage(ctx, update.Message)
			}
		}
	}
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./src/modules/telegram/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add src/modules/telegram/telegram.bot.go
git commit -m "feat(telegram): add bot setup and update loop"
```

---

### Task 9: Telegram Wire Provider

**Files:**
- Create: `src/modules/telegram/telegram.provider.go`

- [ ] **Step 1: Write the provider**

`src/modules/telegram/telegram.provider.go`:
```go
package telegram

import "github.com/google/wire"

// ProviderSet binds telegram constructors for Wire.
var ProviderSet = wire.NewSet(
	NewCache,
	NewBot,
)
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./src/modules/telegram/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add src/modules/telegram/telegram.provider.go
git commit -m "feat(telegram): add wire provider set"
```

---

### Task 10: Wire initializer and main

**Files:**
- Create: `wire.go`
- Create: `wire_gen.go`
- Create: `main.go`

- [ ] **Step 1: Write the wire injector declaration**

`wire.go`:
```go
//go:build wireinject

package main

import (
	"github.com/google/wire"

	"telegram-bot/src/modules/telegram"
	"telegram-bot/src/modules/tiktok"
)

func initBot(token, baseURL string) (*telegram.Bot, error) {
	wire.Build(tiktok.ProviderSet, telegram.ProviderSet)
	return nil, nil
}
```

- [ ] **Step 2: Write the generated wire file**

`wire_gen.go`:
```go
// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject

package main

import (
	"telegram-bot/src/modules/telegram"
	"telegram-bot/src/modules/tiktok"
)

func initBot(token, baseURL string) (*telegram.Bot, error) {
	client := tiktok.NewClient(baseURL)
	service := tiktok.NewService(client)
	cache := telegram.NewCache()
	bot, err := telegram.NewBot(token, service, cache)
	if err != nil {
		return nil, err
	}
	return bot, nil
}
```

- [ ] **Step 3: Write main.go**

`main.go`:
```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	baseURL := os.Getenv("TIKWM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://www.tikwm.com"
	}

	bot, err := initBot(token, baseURL)
	if err != nil {
		log.Fatalf("init bot: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("Bot started. Press Ctrl+C to stop.")
	bot.Start(ctx)
	log.Println("Bot stopped.")
}
```

- [ ] **Step 4: Build the whole project**

```bash
go build ./...
```

Expected: produces a `telegram-bot` binary with no errors.

- [ ] **Step 5: Run all tests**

```bash
go test ./... -v
```

Expected: all tests PASS.

- [ ] **Step 6: Commit**

```bash
git add wire.go wire_gen.go main.go
git commit -m "feat: wire up bot and add main entry point"
```

---

## Self-review

**Spec coverage:**
- ✅ TikTok DTO with `Images` and `IsImage()` — Task 2
- ✅ tikwm HTTP client (`hdplay` → `play` fallback) — Task 3
- ✅ `Service` interface + validation — Task 4
- ✅ Wire provider sets — Tasks 5 and 9
- ✅ Short-ID TTL cache (1h, janitor every 10m) — Task 6
- ✅ `OnMessage`: `/start`, non-TikTok guard, video send, slideshow send, MP3 button — Task 7
- ✅ `OnCallback`: audio send, expired-ID fallback — Task 7
- ✅ MP3 button omitted when `Music` is empty — Task 7 (`mp3Button` returns nil)
- ✅ Media group chunked to ≤10 — Task 7 (`sendImages`)
- ✅ Bot long-polling loop with context cancellation — Task 8
- ✅ `.env` loading, required token check, `TIKWM_BASE_URL` default — Task 10

**Placeholder scan:** None found.

**Type consistency:**
- `tiktok.Service` defined in Task 4, consumed in Tasks 7 and 8 ✅
- `telegram.Sender` interface defined in Task 7; `*tgbotapi.BotAPI` satisfies it (`Send`, `Request`, `SendMediaGroup` all present) ✅
- `telegram.Cache` defined in Task 6, used in Tasks 7 and 8 ✅
- `tiktok.NewClient(baseURL string)` matches `wire_gen.go` call ✅
- `telegram.NewBot(token string, service tiktok.Service, cache *Cache)` matches `wire_gen.go` call ✅
