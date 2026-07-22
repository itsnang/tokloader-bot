//go:build wireinject

package app

import (
	"github.com/google/wire"

	"telegram-bot/src/modules/telegram"
	"telegram-bot/src/modules/tiktok"
)

func InitBot(token, baseURL string) (*telegram.Bot, error) {
	wire.Build(tiktok.ProviderSet, telegram.ProviderSet)
	return nil, nil
}
