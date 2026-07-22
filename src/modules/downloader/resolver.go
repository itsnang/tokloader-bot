package downloader

import (
	"context"
	"errors"
)

// Resolver resolves a platform URL into downloadable media.
type Resolver interface {
	Resolve(ctx context.Context, url string) (*MediaResponse, error)
}

// ErrUnsupportedURL is returned when no provider matches the given URL.
var ErrUnsupportedURL = errors.New("unsupported link — send a TikTok, Instagram, or Facebook URL")
