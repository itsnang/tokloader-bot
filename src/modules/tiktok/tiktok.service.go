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
