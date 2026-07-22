//go:build wireinject

package main

import (
	"github.com/google/wire"

	"telegram-bot/src/modules/downloader"
	"telegram-bot/src/modules/telegram"
)

func initBot(token string, cfg downloader.Config) (*telegram.Bot, error) {
	wire.Build(downloader.ProviderSet, telegram.ProviderSet)
	return nil, nil
}
