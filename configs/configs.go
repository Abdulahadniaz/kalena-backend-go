package config

import (
	"os"
)

type Config struct {
	ServerPort           string
	OAuthCredentialsPath string
}

func LoadConfig() *Config {
	// Check for environment variable, fallback to default if not set
	oauthCredPath := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if oauthCredPath == "" {
		oauthCredPath = "configs/credentials.json"
	}

	return &Config{
		ServerPort:           ":8080",
		OAuthCredentialsPath: oauthCredPath,
	}
}
