package tiktok

import "github.com/google/wire"

// ProviderSet binds tiktok constructors for Wire.
var ProviderSet = wire.NewSet(
	NewClient,
	wire.Bind(new(InfoClient), new(*Client)),
	NewService,
)
