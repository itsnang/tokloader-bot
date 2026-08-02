package downloader

import "context"

type InstagramService struct {
	client *CobaltClient
}

func NewInstagramService(client *CobaltClient) *InstagramService {
	return &InstagramService{client: client}
}

func (s *InstagramService) Resolve(ctx context.Context, rawURL string) (*MediaResponse, error) {
	cr, err := s.client.Fetch(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	return cr.toMediaResponse(), nil
}
