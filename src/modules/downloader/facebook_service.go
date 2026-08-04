package downloader

import (
	"context"
	"fmt"
	"strings"
)

type FacebookService struct {
	client *CobaltClient
}

func NewFacebookService(client *CobaltClient) *FacebookService {
	return &FacebookService{client: client}
}

func (s *FacebookService) Resolve(ctx context.Context, rawURL string) (*MediaResponse, error) {
	if strings.Contains(rawURL, "/stories/") {
		return nil, fmt.Errorf("Facebook stories are not supported: %w", ErrUnsupportedURL)
	}
	cr, err := s.client.Fetch(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	return cr.toMediaResponse(), nil
}
