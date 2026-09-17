// Package config regroupe la configuration de l'application.
// Toutes les valeurs sont lues depuis les variables d'environnement afin
// qu'aucun secret ne soit écrit en dur dans le code source.
package config

import (
	"errors"
	"os"
	"time"
)

// Config contient l'ensemble des paramètres nécessaires au démarrage du serveur.
type Config struct {
	Port          string
	DatabaseURL   string
	JWTSecret     string
	JWTExpiration time.Duration
}

// Load lit la configuration depuis l'environnement.
// Elle retourne une erreur si une valeur obligatoire est absente : il vaut
// mieux refuser de démarrer que de tourner avec une configuration incomplète.
func Load() (*Config, error) {
	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTExpiration: 24 * time.Hour,
	}

	if cfg.DatabaseURL == "" {
		return nil, errors.New("la variable d'environnement DATABASE_URL est obligatoire")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("la variable d'environnement JWT_SECRET est obligatoire")
	}

	return cfg, nil
}

// getEnv retourne la valeur de la variable d'environnement demandée,
// ou la valeur par défaut fournie si celle-ci n'est pas définie.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
