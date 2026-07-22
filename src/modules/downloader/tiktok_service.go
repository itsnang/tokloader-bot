package downloader

import (
	"context"
	"fmt"
)

// TikTokService resolves TikTok URLs via the tikwm.com API.
type TikTokService struct {
	client *TikTokClient
}

func NewTikTokService(client *TikTokClient) *TikTokService {
	return &TikTokService{client: client}
}

func (s *TikTokService) Resolve(ctx context.Context, rawURL string) (*MediaResponse, error) {
	media, err := s.client.Info(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	if media.VideoURL == "" && !media.IsImage() {
		return nil, fmt.Errorf("resolver returned no usable video URL")
	}
	return media, nil
}
