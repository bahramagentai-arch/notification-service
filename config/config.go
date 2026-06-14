package config

import (
	"fmt"
	"os"
)

type config struct {
	DatabaseURL string
	HttpAddr    string
}

func LoadConfig() (config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return config{}, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	return config{
		DatabaseURL: dbURL,
		HttpAddr:    httpAddr,
	}, nil
}
