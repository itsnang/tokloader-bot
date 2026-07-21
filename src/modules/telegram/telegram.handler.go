package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/src/modules/tiktok"
)

// Sender is the subset of tgbotapi.BotAPI methods used by the handler.
// Defined as an interface so the handler can be tested without a live bot.
type Sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
	SendMediaGroup(config tgbotapi.MediaGroupConfig) ([]tgbotapi.Message, error)
}

// Handler processes Telegram messages and callback queries.
type Handler struct {
	sender  Sender
	service tiktok.Service
	cache   *Cache
}

// NewHandler creates a Handler.
func NewHandler(sender Sender, service tiktok.Service, cache *Cache) *Handler {
	return &Handler{sender: sender, service: service, cache: cache}
}

// OnMessage handles an incoming text message.
func (h *Handler) OnMessage(ctx context.Context, msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.Text)

	if text == "/start" {
		h.send(tgbotapi.NewMessage(msg.Chat.ID, "Send me a TikTok link and I'll grab it for you 🎬"))
		return
	}
	if !strings.Contains(strings.ToLower(text), "tiktok.com") {
		h.send(tgbotapi.NewMessage(msg.Chat.ID, "That doesn't look like a TikTok link 🤔"))
		return
	}

	wait, _ := h.sender.Send(tgbotapi.NewMessage(msg.Chat.ID, "Processing... ⏳"))
	defer h.sender.Request(tgbotapi.NewDeleteMessage(msg.Chat.ID, wait.MessageID))

	info, err := h.service.Info(ctx, text)
	if err != nil {
		h.send(tgbotapi.NewMessage(msg.Chat.ID, "Couldn't fetch that video 😕 "+err.Error()))
		return
	}

	mp3Btn := h.mp3Button(info)
	caption := buildCaption(info)

	if info.IsImage() {
		h.sendImages(msg.Chat.ID, info.Images)
		out := tgbotapi.NewMessage(msg.Chat.ID, caption)
		if mp3Btn != nil {
			out.ReplyMarkup = mp3Btn
		}
		h.send(out)
	} else {
		video := tgbotapi.NewVideo(msg.Chat.ID, tgbotapi.FileURL(info.NoWatermark))
		video.Caption = caption
		if mp3Btn != nil {
			video.ReplyMarkup = mp3Btn
		}
		h.send(video)
	}
}

// OnCallback handles the MP3 button tap.
func (h *Handler) OnCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	defer h.sender.Request(tgbotapi.NewCallback(cb.ID, ""))

	if !strings.HasPrefix(cb.Data, "mp3:") {
		return
	}
	id := strings.TrimPrefix(cb.Data, "mp3:")

	musicURL, title, ok := h.cache.Get(id)
	if !ok || musicURL == "" {
		h.send(tgbotapi.NewMessage(cb.Message.Chat.ID, "That MP3 link expired ⌛ send the video again."))
		return
	}

	audio := tgbotapi.NewAudio(cb.Message.Chat.ID, tgbotapi.FileURL(musicURL))
	if title != "" {
		audio.Title = title
	}
	h.send(audio)
}

func (h *Handler) mp3Button(info *tiktok.InfoResponse) *tgbotapi.InlineKeyboardMarkup {
	if info.Music == "" {
		return nil
	}
	id := h.cache.Put(info.Music, info.Title)
	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎵 Download MP3", "mp3:"+id),
		),
	)
	return &markup
}

func (h *Handler) sendImages(chatID int64, images []string) {
	media := make([]interface{}, 0, len(images))
	for _, img := range images {
		media = append(media, tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(img)))
	}
	for i := 0; i < len(media); i += 10 {
		end := i + 10
		if end > len(media) {
			end = len(media)
		}
		h.sender.SendMediaGroup(tgbotapi.NewMediaGroup(chatID, media[i:end]))
	}
}

func (h *Handler) send(c tgbotapi.Chattable) {
	h.sender.Send(c) //nolint:errcheck
}

func buildCaption(info *tiktok.InfoResponse) string {
	if info.Author != "" {
		return fmt.Sprintf("%s\n\n👤 %s", info.Title, info.Author)
	}
	return info.Title
}
