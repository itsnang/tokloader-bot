package downloader

import "github.com/google/wire"

// ProviderSet binds all downloader constructors for Wire.
var ProviderSet = wire.NewSet(
	NewTikTokClient,
	NewTikTokService,
	NewCobaltClient,
	NewInstagramService,
	NewFacebookService,
	NewRouter,
	wire.Bind(new(Resolver), new(*Router)),
)
