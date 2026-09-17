// Package service contient la logique métier de l'application.
//
// C'est ici que vivent les règles : normalisation des données, vérification
// des mots de passe, contrôle d'appartenance des ressources. Les services ne
// connaissent ni le protocole HTTP ni le SQL ; ils orchestrent les
// repositories et le package auth.
package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
	"github.com/KxroTM/test-technique-ynov/internal/auth"
	"github.com/KxroTM/test-technique-ynov/internal/models"
	"github.com/KxroTM/test-technique-ynov/internal/repository"
)

// AuthService gère l'inscription et la connexion des utilisateurs.
type AuthService struct {
	users  *repository.UserRepository
	tokens *auth.TokenManager
}

// NewAuthService construit le service d'authentification.
func NewAuthService(users *repository.UserRepository, tokens *auth.TokenManager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

// Register crée un nouvel utilisateur.
//
// Le mot de passe en clair ne quitte pas cette fonction : il est haché
// immédiatement et seule l'empreinte est transmise au repository.
func (s *AuthService) Register(ctx context.Context, email, plainPassword, name string) (*models.User, error) {
	email = normalizeEmail(email)
	name = strings.TrimSpace(name)

	passwordHash, err := auth.HashPassword(plainPassword)
	if err != nil {
		return nil, err
	}

	return s.users.Create(ctx, email, passwordHash, name)
}

// Login vérifie les identifiants fournis et retourne un jeton JWT.
//
// En cas d'échec, la même erreur est retournée que l'email soit inconnu ou
// que le mot de passe soit faux. Distinguer les deux cas permettrait à un
// attaquant de savoir quelles adresses sont enregistrées.
func (s *AuthService) Login(ctx context.Context, email, plainPassword string) (string, time.Time, *models.User, error) {
	user, err := s.users.GetByEmail(ctx, normalizeEmail(email))
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return "", time.Time{}, nil, apperrors.ErrInvalidCredentials
		}
		return "", time.Time{}, nil, err
	}

	if !auth.CheckPassword(user.PasswordHash, plainPassword) {
		return "", time.Time{}, nil, apperrors.ErrInvalidCredentials
	}

	token, expiresAt, err := s.tokens.Generate(user.ID)
	if err != nil {
		return "", time.Time{}, nil, err
	}

	return token, expiresAt, user, nil
}

// GetUserByID retourne l'utilisateur correspondant à l'identifiant fourni.
// Elle est utilisée pour alimenter l'endpoint « profil courant ».
func (s *AuthService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return s.users.GetByID(ctx, id)
}

// normalizeEmail met l'adresse en minuscules et retire les espaces autour.
//
// Cette normalisation est appliquée à l'inscription comme à la connexion :
// « Alice@Example.com » et « alice@example.com » désignent ainsi le même
// compte, ce qui évite les doublons et les échecs de connexion incompris.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
