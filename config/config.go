package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Env         string `envconfig:"ENV" default:"development"`
	BotApiToken string `envconfig:"TELEGRAM_BOT_API_TOKEN" required:"true"`
	BotUsername string `envconfig:"BOT_USERNAME" required:"true"`
}

var C Config

func Load() {
	if os.Getenv("ENV") != "production" {
		_ = godotenv.Load()
	}
	if err := envconfig.Process("", &C); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
}
