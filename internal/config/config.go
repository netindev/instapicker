package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Username string
	Password string
	Headless bool
	URL      string
}

func Load() Config {
	godotenv.Load()
	return Config{
		Username: os.Getenv("INSTAGRAM_USERNAME"),
		Password: os.Getenv("INSTAGRAM_PASSWORD"),
		Headless: os.Getenv("HEADLESS") != "false",
		URL:      os.Getenv("INSTAGRAM_URL"),
	}
}
