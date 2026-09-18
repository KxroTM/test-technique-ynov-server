// Package config regroupe la configuration de l'application
package config

import (
	"errors"
	"os"
	"time"
)

// Config contient l'ensemble des paramètres nécessaires au démarrage du serveur
type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	JWTExpiration      time.Duration
	GoogleClientID     string
	GoogleClientSecret string
}

// Load lit la configuration depuis l'environnement
func Load() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		JWTExpiration:      24 * time.Hour,
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("la variable d'environnement DATABASE_URL est obligatoire")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("la variable d'environnement JWT_SECRET est obligatoire")
	}

	return cfg, nil
}

// GoogleEnabled indique si la connexion Google est configurée, l'application restant pleinement utilisable sans elle
func (c *Config) GoogleEnabled() bool {
	return c.GoogleClientID != "" && c.GoogleClientSecret != ""
}

// getEnv retourne la valeur de la variable d'environnement demandée, ou la valeur par défaut fournie si celle-ci n'est pas définie
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
