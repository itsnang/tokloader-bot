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
