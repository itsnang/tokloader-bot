package downloader

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewTikTokClient,
	NewTikTokService,
	NewCobaltClient,
	NewInstagramService,
	NewFacebookService,
	NewYouTubeService,
	NewRouter,
	wire.Bind(new(Resolver), new(*Router)),
)
