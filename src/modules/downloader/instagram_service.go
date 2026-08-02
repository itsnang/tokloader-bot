package downloader

import "context"

// InstagramService resolves Instagram reels, posts, and stories via cobalt.
// Story support requires INSTAGRAM_COOKIE to be configured in the cobalt service (not the bot).
type InstagramService struct {
	client *CobaltClient
}

func NewInstagramService(client *CobaltClient) *InstagramService {
	return &InstagramService{client: client}
}

func (s *InstagramService) Resolve(ctx context.Context, rawURL string) (*MediaResponse, error) {
	cr, err := s.client.Fetch(ctx, rawURL, "auto", "")
	if err != nil {
		return nil, err
	}
	return cr.toMediaResponse(), nil
}
