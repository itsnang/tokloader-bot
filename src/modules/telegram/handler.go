package telegram

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"telegram-bot/src/modules/downloader"
)

const callbackPrefixMP3 = "mp3:"

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
	service downloader.Resolver
	cache   *Cache
}

func NewHandler(sender Sender, service downloader.Resolver, cache *Cache) *Handler {
	return &Handler{sender: sender, service: service, cache: cache}
}

func (h *Handler) OnMessage(ctx context.Context, msg *tgbotapi.Message) {
	text := strings.TrimSpace(msg.Text)

	if text == "/start" {
		h.send(tgbotapi.NewMessage(msg.Chat.ID, "Send me a TikTok, Instagram, or Facebook link and I'll grab it for you 🎬"))
		return
	}
	if !strings.HasPrefix(text, "http://") && !strings.HasPrefix(text, "https://") {
		h.send(tgbotapi.NewMessage(msg.Chat.ID, "Send me a link from TikTok, Instagram, or Facebook 🔗"))
		return
	}

	wait, err := h.sender.Send(tgbotapi.NewMessage(msg.Chat.ID, "Processing... ⏳"))
	if err == nil {
		defer h.sender.Request(tgbotapi.NewDeleteMessage(msg.Chat.ID, wait.MessageID))
	}

	info, err := h.service.Resolve(ctx, text)
	if err != nil {
		if errors.Is(err, downloader.ErrUnsupportedURL) {
			h.send(tgbotapi.NewMessage(msg.Chat.ID, "Unsupported link 🤔 Send a TikTok, Instagram, or Facebook URL."))
		} else {
			h.send(tgbotapi.NewMessage(msg.Chat.ID, "Couldn't fetch that video 😕 "+err.Error()))
		}
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
		video := tgbotapi.NewVideo(msg.Chat.ID, tgbotapi.FileURL(info.VideoURL))
		video.Caption = caption
		if mp3Btn != nil {
			video.ReplyMarkup = mp3Btn
		}
		h.send(video)
	}
}

func (h *Handler) OnCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	defer h.sender.Request(tgbotapi.NewCallback(cb.ID, ""))

	if cb.Message == nil {
		return
	}

	prefix, id, ok := strings.Cut(cb.Data, ":")
	if !ok || prefix != "mp3" {
		return
	}

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

func (h *Handler) mp3Button(info *downloader.MediaResponse) *tgbotapi.InlineKeyboardMarkup {
	if info.AudioURL == "" {
		return nil
	}
	id := h.cache.Put(info.AudioURL, info.Title)
	markup := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🎵 Download MP3", callbackPrefixMP3+id),
		),
	)
	return &markup
}

func (h *Handler) sendImages(chatID int64, images []string) {
	// tgbotapi.NewMediaGroup requires []interface{} — library predates generics.
	media := make([]interface{}, 0, len(images))
	for _, img := range images {
		media = append(media, tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(img)))
	}
	for i := 0; i < len(media); i += 10 {
		end := i + 10
		if end > len(media) {
			end = len(media)
		}
		if _, err := h.sender.SendMediaGroup(tgbotapi.NewMediaGroup(chatID, media[i:end])); err != nil {
			log.Printf("sendImages: SendMediaGroup failed: %v", err)
		}
	}
}

func (h *Handler) send(c tgbotapi.Chattable) {
	if _, err := h.sender.Send(c); err != nil {
		log.Printf("send failed: %v", err)
	}
}

func buildCaption(info *downloader.MediaResponse) string {
	if info.Author != "" {
		return fmt.Sprintf("%s\n\n👤 %s", info.Title, info.Author)
	}
	return info.Title
}
