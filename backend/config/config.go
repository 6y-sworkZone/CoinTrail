package config

import "os"

type Config struct {
	Port        string
	JWTSecret   string
	DatabaseURL string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8765"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "cointrail-secret-key-2024"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "cointrail.db"
	}

	return &Config{
		Port:        port,
		JWTSecret:   jwtSecret,
		DatabaseURL: dbURL,
	}
}
