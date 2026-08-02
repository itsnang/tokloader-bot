package downloader

import (
	"context"
	"fmt"
	"net/http"
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
	target := rawURL
	if strings.Contains(rawURL, "/share/") {
		if final, err := resolveRedirect(ctx, rawURL); err == nil {
			target = final
		}
	}
	cr, err := s.client.Fetch(ctx, target)
	if err != nil {
		return nil, err
	}
	return cr.toMediaResponse(), nil
}

func resolveRedirect(ctx context.Context, rawURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	return resp.Request.URL.String(), nil
}
