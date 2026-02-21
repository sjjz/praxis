package lib

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

const defaultHTTPAddr = ":8080"

type Config struct {
	HTTPAddr   string
	DatabaseURL string
	DevUserID  uuid.UUID
}

func LoadConfig() (Config, error) {
	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = defaultHTTPAddr
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	devUser := os.Getenv("DEV_USER_ID")
	if devUser == "" {
		return Config{}, fmt.Errorf("DEV_USER_ID is required")
	}
	devUserID, err := uuid.Parse(devUser)
	if err != nil {
		return Config{}, fmt.Errorf("DEV_USER_ID must be a valid UUID: %w", err)
	}

	return Config{
		HTTPAddr:   httpAddr,
		DatabaseURL: dbURL,
		DevUserID:  devUserID,
	}, nil
}
