package config

import (
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

// Config содержит конфигурацию приложения
type Config struct {
	ServiceName string `env:"SERVICE_NAME" required:"true" default:"account-service"`
	AppEnv      string `env:"APP_ENV" required:"true" default:"development"`
	Host        string `env:"HTTP_HOST" required:"true" default:"localhost"`
	Port        int    `env:"HTTP_PORT" required:"true" default:"50052"`
	LogLevel    string `env:"LOG_LEVEL" required:"true" default:"info"`
	DbDsn       string `env:"DB_DSN" required:"true"`

	JwtSecret             string `env:"JWT_SECRET" json:"jwt_secret" required:"true"`
	AccessTokenTTLMinutes int    `env:"ACCESS_TOKEN_TTL_MINUTES" json:"access_ttl_min" required:"true" default:"60"`
	RefreshTokenTTLDays   int    `env:"REFRESH_TOKEN_TTL_DAYS" json:"refresh_ttl_days" required:"true" default:"30"`
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {

	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
