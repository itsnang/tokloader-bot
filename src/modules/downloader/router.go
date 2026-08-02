package downloader

import (
	"context"
	"net/url"
	"strings"
)

type Router struct {
	tiktok    Resolver
	instagram Resolver
	facebook  Resolver
}

func NewRouter(t *TikTokService, i *InstagramService, f *FacebookService) *Router {
	return &Router{tiktok: t, instagram: i, facebook: f}
}

func NewRouterWithResolvers(tiktok, instagram, facebook Resolver) *Router {
	return &Router{tiktok: tiktok, instagram: instagram, facebook: facebook}
}

func (r *Router) Resolve(ctx context.Context, rawURL string) (*MediaResponse, error) {
	host, err := parseHost(rawURL)
	if err != nil {
		return nil, ErrUnsupportedURL
	}
	switch {
	case strings.Contains(host, "tiktok.com"):
		return r.tiktok.Resolve(ctx, rawURL)
	case strings.Contains(host, "instagram.com"):
		return r.instagram.Resolve(ctx, rawURL)
	case strings.Contains(host, "facebook.com"), strings.Contains(host, "fb.watch"):
		return r.facebook.Resolve(ctx, rawURL)
	}
	return nil, ErrUnsupportedURL
}

func parseHost(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return "", ErrUnsupportedURL
	}
	return strings.ToLower(u.Host), nil
}
