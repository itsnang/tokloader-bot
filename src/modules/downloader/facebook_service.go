package downloader

import (
	"context"
	"fmt"
	"strings"
)

// FacebookService resolves Facebook videos and stories via cobalt.tools.
//
// Stories require a logged-in session cookie. Without FACEBOOK_COOKIE set,
// story requests return a descriptive error rather than silently failing.
type FacebookService struct {
	client *CobaltClient
	cookie string
}

func NewFacebookService(client *CobaltClient, cfg Config) *FacebookService {
	return &FacebookService{client: client, cookie: cfg.FacebookCookie}
}

func (s *FacebookService) Resolve(ctx context.Context, rawURL string) (*MediaResponse, error) {
	if strings.Contains(rawURL, "/stories/") && s.cookie == "" {
		return nil, fmt.Errorf("Facebook stories require a session cookie — set FACEBOOK_COOKIE in config")
	}
	cr, err := s.client.Fetch(ctx, rawURL, "auto", s.cookie)
	if err != nil {
		return nil, err
	}
	return cr.toMediaResponse(), nil
}
