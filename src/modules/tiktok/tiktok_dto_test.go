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
