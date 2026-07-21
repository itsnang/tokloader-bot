//go:build wireinject

package main

import (
	"github.com/google/wire"

	"telegram-bot/src/modules/telegram"
	"telegram-bot/src/modules/tiktok"
)

func initBot(token, baseURL string) (*telegram.Bot, error) {
	wire.Build(tiktok.ProviderSet, telegram.ProviderSet)
	return nil, nil
}
