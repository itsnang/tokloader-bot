package downloader

import "context"

type YouTubeService struct {
	client *CobaltClient
}

func NewYouTubeService(client *CobaltClient) *YouTubeService {
	return &YouTubeService{client: client}
}

func (s *YouTubeService) Resolve(ctx context.Context, rawURL string) (*MediaResponse, error) {
	cr, err := s.client.Fetch(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	return cr.toMediaResponse(), nil
}
