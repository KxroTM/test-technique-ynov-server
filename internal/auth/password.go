// Package auth regroupe le hachage des mots de passe et la gestion des jetons JWT

package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword calcule l'empreinte bcrypt d'un mot de passe en clair
func HashPassword(plainPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hachage du mot de passe : %w", err)
	}
	return string(hash), nil
}

// CheckPassword vérifie qu'un mot de passe en clair correspond à une empreinte
func CheckPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
