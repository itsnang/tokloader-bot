package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/src/modules/downloader"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	handler *Handler
}

func NewBot(token string, service downloader.Resolver, cache *Cache) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	handler := NewHandler(api, service, cache)
	return &Bot{api: api, handler: handler}, nil
}

// Start runs the long-polling update loop. Blocks until ctx is cancelled.
func (b *Bot) Start(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			b.api.StopReceivingUpdates()
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			switch {
			case update.CallbackQuery != nil:
				go b.handler.OnCallback(ctx, update.CallbackQuery)
			case update.Message != nil && update.Message.Text != "":
				go b.handler.OnMessage(ctx, update.Message)
			}
		}
	}
}
