package downloader

import (
	"context"
	"fmt"
	"strings"
)

// InstagramService resolves Instagram reels, posts, and stories via cobalt.tools.
//
// Stories require a logged-in session cookie. Without INSTAGRAM_COOKIE set,
// story requests return a descriptive error rather than silently failing.
type InstagramService struct {
	client *CobaltClient
	cookie string
}

func NewInstagramService(client *CobaltClient, cfg Config) *InstagramService {
	return &InstagramService{client: client, cookie: cfg.InstagramCookie}
}

func (s *InstagramService) Resolve(ctx context.Context, rawURL string) (*MediaResponse, error) {
	if strings.Contains(rawURL, "/stories/") && s.cookie == "" {
		return nil, fmt.Errorf("Instagram stories require a session cookie — set INSTAGRAM_COOKIE in config")
	}
	cr, err := s.client.Fetch(ctx, rawURL, "auto", s.cookie)
	if err != nil {
		return nil, err
	}
	return cr.toMediaResponse(), nil
}
