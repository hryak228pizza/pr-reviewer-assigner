// Package config loads application configuration
package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config holds all application configuration settings
type Config struct {
	Env        string `env:"ENV" env-default:"local"`
	HTTPServer HTTPServer
	PG         PG
}

// HTTPServer holds http server-specific configuration
type HTTPServer struct {
	Address     string        `env:"HTTP_ADDRESS" env-default:"localhost:8080"`
	Timeout     time.Duration `env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

// PG holds postgresql database configuration
type PG struct {
	URL string `env:"PG_URL" env-required:"true"`
}

// Load reads configuration from env file and environment variables
// it uses a fatal log if required variables are missing
func Load() *Config {
	cfg := &Config{}

	// read env file first (non-fatal if missing)
	if err := cleanenv.ReadConfig(".env", cfg); err != nil {
		log.Printf("INFO: .env file not found or failed to read: %v", err)
	}

	// read environment variables (fatal if required variables are missing)
	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("FATAL: cannot read environment configuration: %v", err)
	}

	return cfg
}
