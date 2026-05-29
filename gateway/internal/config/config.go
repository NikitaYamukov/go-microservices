package config

import (
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

// Config содержит конфигурацию приложения.
type Config struct {
	ServiceName string `env:"SERVICE_NAME" json:"service_name" required:"true" default:"account-service"`
	AppEnv      string `env:"APP_ENV" json:"app_environment" required:"true" default:"development"`
	Host        string `env:"GRPC_HOST" json:"host" required:"true" default:"localhost"`
	Port        int    `env:"GRPC_PORT" json:"port" required:"true" default:"50053"`
	LogLevel    string `env:"LOG_LEVEL" json:"log_level" required:"true" default:"info"`

	AccountGrpcHost string `env:"ACCOUNT_GRPC_HOST" json:"account_grpc_host" required:"true"`
	AuthGrpcHost    string `env:"AUTH_GRPC_HOST" json:"auth_grpc_host" required:"true"`
	JwtSecret       string `env:"JWT_SECRET" json:"jwt_secret" required:"true"`
}

// Load загружает конфигурацию из переменных окружения.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
