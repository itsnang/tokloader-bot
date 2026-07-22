package tiktok

import (
	"context"
	"fmt"
)

// InfoClient is satisfied by *Client but defined here so service.go has no import cycle.
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
