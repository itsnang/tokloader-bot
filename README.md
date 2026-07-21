# tokloader-bot

A Telegram bot that downloads TikTok videos and slideshows — watermark-free.

## Features

- Send a TikTok link → get back a watermark-free video
- Slideshow posts → receives all images as a photo group
- Inline **🎵 Download MP3** button to grab the background audio

## Requirements

- Go 1.22+
- A Telegram bot token (from [@BotFather](https://t.me/BotFather))

## Setup

```bash
cp .env.example .env
# Edit .env and fill in your TELEGRAM_BOT_TOKEN
```

## Run

```bash
go run .
```

## Configuration

| Variable | Required | Default |
|---|---|---|
| `TELEGRAM_BOT_TOKEN` | yes | — |
| `TIKWM_BASE_URL` | no | `https://www.tikwm.com` |

## Test

```bash
go test ./...
```

## Project Structure

```
telegram-bot/
├── main.go
├── wire_gen.go
└── src/modules/
    ├── tiktok/        # tikwm.com API client + service
    └── telegram/      # bot, handler, TTL cache
```
