package downloader_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"telegram-bot/src/modules/downloader"
)

func newTikwmServer(t *testing.T, response map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

func TestTikTokClient_Info_Video(t *testing.T) {
	srv := newTikwmServer(t, map[string]any{
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
			"music":    "https://example.com/music.mp3",
		},
	})
	defer srv.Close()

	client := downloader.NewTikTokClient(downloader.Config{TikwmBaseURL: srv.URL})
	info, err := client.Info(context.Background(), "https://tiktok.com/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.ID != "123" {
		t.Errorf("ID: want 123, got %s", info.ID)
	}
	if info.VideoURL != "https://example.com/video_hd.mp4" {
		t.Errorf("VideoURL: want hdplay URL, got %s", info.VideoURL)
	}
	if info.AudioURL != "https://example.com/music.mp3" {
		t.Errorf("AudioURL: want music URL, got %s", info.AudioURL)
	}
	if info.IsImage() {
		t.Error("expected IsImage() = false for video post")
	}
}

func TestTikTokClient_Info_HDPlayFallback(t *testing.T) {
	srv := newTikwmServer(t, map[string]any{
		"code": 0,
		"msg":  "success",
		"data": map[string]any{
			"id":     "124",
			"title":  "No HD video",
			"author": map[string]any{"unique_id": "testuser"},
			"hdplay": "",
			"play":   "https://example.com/video.mp4",
			"music":  "https://example.com/music.mp3",
		},
	})
	defer srv.Close()

	client := downloader.NewTikTokClient(downloader.Config{TikwmBaseURL: srv.URL})
	info, err := client.Info(context.Background(), "https://tiktok.com/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.VideoURL != "https://example.com/video.mp4" {
		t.Errorf("VideoURL: want play fallback URL, got %s", info.VideoURL)
	}
}

func TestTikTokClient_Info_Slideshow(t *testing.T) {
	srv := newTikwmServer(t, map[string]any{
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

	client := downloader.NewTikTokClient(downloader.Config{TikwmBaseURL: srv.URL})
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

func TestTikTokClient_Info_APIError(t *testing.T) {
	srv := newTikwmServer(t, map[string]any{
		"code": -1,
		"msg":  "video not found",
	})
	defer srv.Close()

	client := downloader.NewTikTokClient(downloader.Config{TikwmBaseURL: srv.URL})
	_, err := client.Info(context.Background(), "https://tiktok.com/test")
	if err == nil {
		t.Error("expected error for non-zero API code")
	}
}
