package downloader

import "context"

type TikTokService struct {
	client *TikTokClient
}

func NewTikTokService(client *TikTokClient) *TikTokService {
	return &TikTokService{client: client}
}

func (s *TikTokService) Resolve(ctx context.Context, rawURL string) (*MediaResponse, error) {
	return s.client.Info(ctx, rawURL)
}
