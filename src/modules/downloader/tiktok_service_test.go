package downloader_test

import (
	"context"
	"testing"

	"telegram-bot/src/modules/downloader"
)

func TestTikTokService_Resolve_Success(t *testing.T) {
	srv := newTikwmServer(t, map[string]any{
		"code": 0,
		"msg":  "success",
		"data": map[string]any{
			"id":     "1",
			"title":  "Test",
			"author": map[string]any{"unique_id": "user"},
			"hdplay": "https://cdn.example.com/v.mp4",
			"music":  "https://cdn.example.com/m.mp3",
		},
	})
	defer srv.Close()

	svc := downloader.NewTikTokService(downloader.NewTikTokClient(downloader.Config{TikwmBaseURL: srv.URL}))
	media, err := svc.Resolve(context.Background(), "https://tiktok.com/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if media.VideoURL == "" {
		t.Error("expected non-empty VideoURL")
	}
}

func TestTikTokService_Resolve_EmptyVideoURL(t *testing.T) {
	srv := newTikwmServer(t, map[string]any{
		"code": 0,
		"msg":  "success",
		"data": map[string]any{
			"id":     "2",
			"title":  "Empty",
			"author": map[string]any{"unique_id": "user"},
			"hdplay": "",
			"play":   "",
			"music":  "https://cdn.example.com/m.mp3",
		},
	})
	defer srv.Close()

	svc := downloader.NewTikTokService(downloader.NewTikTokClient(downloader.Config{TikwmBaseURL: srv.URL}))
	_, err := svc.Resolve(context.Background(), "https://tiktok.com/test")
	if err == nil {
		t.Error("expected error when VideoURL is empty and not an image")
	}
}

func TestTikTokService_Resolve_SlideshowNoVideoOK(t *testing.T) {
	srv := newTikwmServer(t, map[string]any{
		"code": 0,
		"msg":  "success",
		"data": map[string]any{
			"id":     "3",
			"title":  "Slideshow",
			"author": map[string]any{"unique_id": "user"},
			"images": []string{"https://example.com/img.jpg"},
			"music":  "https://example.com/music.mp3",
		},
	})
	defer srv.Close()

	svc := downloader.NewTikTokService(downloader.NewTikTokClient(downloader.Config{TikwmBaseURL: srv.URL}))
	media, err := svc.Resolve(context.Background(), "https://tiktok.com/test")
	if err != nil {
		t.Fatalf("unexpected error for slideshow: %v", err)
	}
	if !media.IsImage() {
		t.Error("expected IsImage() = true")
	}
}
