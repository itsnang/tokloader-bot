package downloader

import (
	"context"
	"fmt"
	"strings"
)

// FacebookService resolves Facebook videos via cobalt.
// Facebook stories are not supported by cobalt (the /stories/ URL format is not matched).
type FacebookService struct {
	client *CobaltClient
}

func NewFacebookService(client *CobaltClient) *FacebookService {
	return &FacebookService{client: client}
}

func (s *FacebookService) Resolve(ctx context.Context, rawURL string) (*MediaResponse, error) {
	if strings.Contains(rawURL, "/stories/") {
		return nil, fmt.Errorf("Facebook stories are not supported")
	}
	cr, err := s.client.Fetch(ctx, rawURL, "auto", "")
	if err != nil {
		return nil, err
	}
	return cr.toMediaResponse(), nil
}
