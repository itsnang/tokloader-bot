package downloader_test

import (
	"context"
	"errors"
	"testing"

	"telegram-bot/src/modules/downloader"
)

type stubResolver struct {
	called bool
	media  *downloader.MediaResponse
	err    error
}

func (s *stubResolver) Resolve(_ context.Context, _ string) (*downloader.MediaResponse, error) {
	s.called = true
	return s.media, s.err
}

func TestRouter_TikTok(t *testing.T) {
	tt := &stubResolver{media: &downloader.MediaResponse{VideoURL: "https://cdn.example.com/v.mp4"}}
	ig := &stubResolver{}
	fb := &stubResolver{}
	yt := &stubResolver{}
	r := downloader.NewRouterWithResolvers(tt, ig, fb, yt)

	_, err := r.Resolve(context.Background(), "https://www.tiktok.com/@user/video/1")
	if err != nil {
		t.Fatal(err)
	}
	if !tt.called {
		t.Error("expected tiktok resolver to be called")
	}
	if ig.called || fb.called || yt.called {
		t.Error("only tiktok resolver should be called")
	}
}

func TestRouter_ShortTikTok(t *testing.T) {
	tt := &stubResolver{media: &downloader.MediaResponse{VideoURL: "https://cdn.example.com/v.mp4"}}
	r := downloader.NewRouterWithResolvers(tt, &stubResolver{}, &stubResolver{}, &stubResolver{})

	_, err := r.Resolve(context.Background(), "https://vt.tiktok.com/abcdef/")
	if err != nil {
		t.Fatal(err)
	}
	if !tt.called {
		t.Error("expected tiktok resolver for vt.tiktok.com")
	}
}

func TestRouter_Instagram(t *testing.T) {
	ig := &stubResolver{media: &downloader.MediaResponse{VideoURL: "https://cdn.example.com/v.mp4"}}
	r := downloader.NewRouterWithResolvers(&stubResolver{}, ig, &stubResolver{}, &stubResolver{})

	_, err := r.Resolve(context.Background(), "https://www.instagram.com/reel/abc123/")
	if err != nil {
		t.Fatal(err)
	}
	if !ig.called {
		t.Error("expected instagram resolver to be called")
	}
}

func TestRouter_Facebook(t *testing.T) {
	fb := &stubResolver{media: &downloader.MediaResponse{VideoURL: "https://cdn.example.com/v.mp4"}}
	r := downloader.NewRouterWithResolvers(&stubResolver{}, &stubResolver{}, fb, &stubResolver{})

	_, err := r.Resolve(context.Background(), "https://www.facebook.com/watch/?v=123456")
	if err != nil {
		t.Fatal(err)
	}
	if !fb.called {
		t.Error("expected facebook resolver to be called")
	}
}

func TestRouter_FbWatch(t *testing.T) {
	fb := &stubResolver{media: &downloader.MediaResponse{VideoURL: "https://cdn.example.com/v.mp4"}}
	r := downloader.NewRouterWithResolvers(&stubResolver{}, &stubResolver{}, fb, &stubResolver{})

	_, err := r.Resolve(context.Background(), "https://fb.watch/abcdef/")
	if err != nil {
		t.Fatal(err)
	}
	if !fb.called {
		t.Error("expected facebook resolver for fb.watch")
	}
}

func TestRouter_YouTube(t *testing.T) {
	yt := &stubResolver{media: &downloader.MediaResponse{VideoURL: "https://cdn.example.com/v.mp4"}}
	r := downloader.NewRouterWithResolvers(&stubResolver{}, &stubResolver{}, &stubResolver{}, yt)

	_, err := r.Resolve(context.Background(), "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	if err != nil {
		t.Fatal(err)
	}
	if !yt.called {
		t.Error("expected youtube resolver to be called")
	}
}

func TestRouter_YouTubeShort(t *testing.T) {
	yt := &stubResolver{media: &downloader.MediaResponse{VideoURL: "https://cdn.example.com/v.mp4"}}
	r := downloader.NewRouterWithResolvers(&stubResolver{}, &stubResolver{}, &stubResolver{}, yt)

	_, err := r.Resolve(context.Background(), "https://youtu.be/dQw4w9WgXcQ")
	if err != nil {
		t.Fatal(err)
	}
	if !yt.called {
		t.Error("expected youtube resolver for youtu.be")
	}
}

func TestRouter_Unsupported(t *testing.T) {
	r := downloader.NewRouterWithResolvers(&stubResolver{}, &stubResolver{}, &stubResolver{}, &stubResolver{})

	_, err := r.Resolve(context.Background(), "https://reddit.com/r/videos/abc")
	if !errors.Is(err, downloader.ErrUnsupportedURL) {
		t.Errorf("expected ErrUnsupportedURL, got %v", err)
	}
}

func TestRouter_InvalidURL(t *testing.T) {
	r := downloader.NewRouterWithResolvers(&stubResolver{}, &stubResolver{}, &stubResolver{}, &stubResolver{})

	_, err := r.Resolve(context.Background(), "not-a-url")
	if !errors.Is(err, downloader.ErrUnsupportedURL) {
		t.Errorf("expected ErrUnsupportedURL, got %v", err)
	}
}
