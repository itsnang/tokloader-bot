package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"telegram-bot/src/app"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	baseURL := os.Getenv("TIKWM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://www.tikwm.com"
	}

	bot, err := app.InitBot(token, baseURL)
	if err != nil {
		log.Fatalf("init bot: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Println("Bot started. Press Ctrl+C to stop.")
	bot.Start(ctx)
	log.Println("Bot stopped.")
}
