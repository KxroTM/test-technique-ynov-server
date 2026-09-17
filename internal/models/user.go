// Package models définit les entités métier de l'application.
//
// Ces structures servent à la fois de représentation interne et de format de
// sérialisation JSON de l'API. Pour une application de cette taille, faire
// porter les deux rôles aux mêmes structures évite une duplication inutile.
package models

import "time"

// User représente un utilisateur de l'application.
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`

	// PasswordHash contient l'empreinte bcrypt du mot de passe.
	// Le tag `json:"-"` garantit que ce champ ne peut jamais être
	// sérialisé dans une réponse de l'API, même par oubli.
	PasswordHash string `json:"-"`
}
