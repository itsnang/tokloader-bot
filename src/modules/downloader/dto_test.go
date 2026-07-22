package downloader_test

import (
	"testing"

	"telegram-bot/src/modules/downloader"
)

func TestIsImage_WithImages(t *testing.T) {
	r := &downloader.MediaResponse{Images: []string{"https://example.com/img1.jpg"}}
	if !r.IsImage() {
		t.Error("expected IsImage() = true when Images is non-empty")
	}
}

func TestIsImage_WithoutImages(t *testing.T) {
	r := &downloader.MediaResponse{}
	if r.IsImage() {
		t.Error("expected IsImage() = false when Images is empty")
	}
}
