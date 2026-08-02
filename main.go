package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"telegram-bot/src/modules/downloader"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	cfg := downloader.Config{
		TikwmBaseURL:    getEnvOrDefault("TIKWM_BASE_URL", "https://www.tikwm.com"),
		CobaltBaseURL:   getEnvOrDefault("COBALT_BASE_URL", "https://api.cobalt.tools"),
		InstagramCookie: os.Getenv("INSTAGRAM_COOKIE"),
		FacebookCookie:  os.Getenv("FACEBOOK_COOKIE"),
	}

	bot, err := initBot(token, cfg)
	if err != nil {
		log.Fatalf("init bot: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("Bot started. Press Ctrl+C to stop.")
	bot.Start(ctx)
	log.Println("Bot stopped.")
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
