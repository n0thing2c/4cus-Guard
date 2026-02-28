package config

import (
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	RedisAddr string
	RedisPass string
	DBPath    string
}

func LoadConfig() *AppConfig {
	_ = godotenv.Load()

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6380"
	}

	pass := os.Getenv("REDIS_PASS")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/focus.db"
	}

	return &AppConfig{
		RedisAddr: addr,
		RedisPass: pass,
		DBPath:    dbPath,
	}
}
