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
