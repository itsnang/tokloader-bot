package downloader

import (
	"context"
	"errors"
)

type Resolver interface {
	Resolve(ctx context.Context, url string) (*MediaResponse, error)
}

var ErrUnsupportedURL = errors.New("unsupported link — send a TikTok, Instagram, or Facebook URL")
