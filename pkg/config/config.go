package config

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port                     int    `envconfig:"PORT" default:"8080" `
	DatabaseURL              string `envconfig:"DATABASE_URL" required:"true"`
	JWTPublicKey             string `envconfig:"JWT_PUBLIC_KEY"`
	JWTSecret                string `envconfig:"JWT_SECRET"`
	MatchmakingBotTimeoutSec int    `envconfig:"MATCHMAKING_BOT_TIMEOUT_SEC" default:"30"`
	TurnTimeoutSec           int    `envconfig:"TURN_TIMEOUT_SEC" default:"45"`
	LogLevel                 string `envconfig:"LOG_LEVEL" default:"info"`
}

type AppConfig struct {
	Config
	JWTPublicKeyParsed *rsa.PublicKey
}

func Load() (*AppConfig, error) {
	// Load .env file and ignore the error if .env is not exist
	_ = godotenv.Overload(".env.dev")

	var cfg Config
	if err := envconfig.Process("DEMONDOOF", &cfg); err != nil {
		return nil, err
	}

	appCfg := &AppConfig{
		Config: cfg,
	}

	// Parse JWT public key if provided
	if cfg.JWTPublicKey != "" {
		// Implementation for parsing JWT public key...
		// For now, just log that it's configured
	}

	// Validate required fields
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// Validate JWT secret is not default in production
	if cfg.JWTSecret == "your-super-secret-jwt-key-change-this-in-production" {
		fmt.Fprintf(os.Stderr, "WARNING: Using default JWT secret in production is insecure\n")
	}

	return appCfg, nil
}
