package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `env:"ENV" env-default:"local"`
	HTTPServer HTTPServer
	PG         PG
}

type HTTPServer struct {
	Address     string        `env:"HTTP_ADDRESS" env-default:"localhost:8080"`
	Timeout     time.Duration `env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type PG struct {
	URL string `env:"PG_URL" env-required:"true"`
}

func Load() *Config {
	cfg := &Config{}

	if err := cleanenv.ReadConfig(".env", cfg); err != nil {
		log.Printf("INFO: .env file not found or failed to read: %v", err)
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		log.Fatalf("FATAL: cannot read environment configuration: %v", err)
	}

	return cfg
}
