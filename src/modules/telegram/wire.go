package telegram

import "github.com/google/wire"

// ProviderSet binds telegram constructors for Wire.
var ProviderSet = wire.NewSet(
	NewCache,
	NewBot,
)
