// Package auth regroupe le hachage des mots de passe et la gestion des
// jetons JWT. Ces deux mécanismes sont isolés du reste de l'application afin
// de pouvoir être testés séparément et remplacés sans toucher à la logique
// métier.
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword calcule l'empreinte bcrypt d'un mot de passe en clair.
//
// bcrypt est utilisé plutôt qu'une fonction de hachage généraliste comme
// SHA-256 : il est volontairement lent et intègre un sel aléatoire, ce qui
// rend les attaques par dictionnaire et par table précalculée coûteuses.
// Le sel est inclus dans la chaîne retournée, il n'y a donc rien d'autre
// à stocker.
func HashPassword(plainPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hachage du mot de passe : %w", err)
	}
	return string(hash), nil
}

// CheckPassword vérifie qu'un mot de passe en clair correspond à une empreinte.
//
// La comparaison est déléguée à bcrypt, qui recalcule l'empreinte avec le sel
// extrait du hash existant puis compare en temps constant.
func CheckPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
