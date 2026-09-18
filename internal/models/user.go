// Package models définit les entités métier de l'application

package models

import "time"

// User représente un utilisateur de l'application
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`

	PasswordHash string `json:"-"`

	// GoogleID est vide pour un compte créé par mot de passe
	GoogleID string `json:"-"`
}

// HasPassword indique si le compte peut se connecter par mot de passe
func (u *User) HasPassword() bool {
	return u.PasswordHash != ""
}
