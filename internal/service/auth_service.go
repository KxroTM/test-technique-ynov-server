// Package service contient la logique métier de l'application

package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
	"github.com/KxroTM/test-technique-ynov/internal/auth"
	"github.com/KxroTM/test-technique-ynov/internal/models"
	"github.com/KxroTM/test-technique-ynov/internal/repository"
)

// AuthService gère l'inscription et la connexion des utilisateurs
type AuthService struct {
	users  *repository.UserRepository
	tokens *auth.TokenManager
	google *auth.GoogleExchanger
}

// NewAuthService construit le service d'authentification
func NewAuthService(users *repository.UserRepository, tokens *auth.TokenManager, google *auth.GoogleExchanger) *AuthService {
	return &AuthService{users: users, tokens: tokens, google: google}
}

// Register crée un nouvel utilisateur
func (s *AuthService) Register(ctx context.Context, email, plainPassword, name string) (*models.User, error) {
	email = normalizeEmail(email)
	name = strings.TrimSpace(name)

	passwordHash, err := auth.HashPassword(plainPassword)
	if err != nil {
		return nil, err
	}

	return s.users.Create(ctx, email, passwordHash, name)
}

// Login vérifie les identifiants fournis et retourne un jeton JWT
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

// GetUserByID retourne l'utilisateur correspondant à l'identifiant fourni
func (s *AuthService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return s.users.GetByID(ctx, id)
}

// normalizeEmail met l'adresse en minuscules et retire les espaces autour
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// LoginWithGoogle authentifie un utilisateur à partir d'un code d'autorisation Google
func (s *AuthService) LoginWithGoogle(ctx context.Context, code, redirectURI string) (string, time.Time, *models.User, error) {
	if s.google == nil {
		return "", time.Time{}, nil, apperrors.ErrGoogleUnavailable
	}

	identity, err := s.google.Exchange(ctx, code, redirectURI)
	if err != nil {
		log.Printf("échange du code Google : %v", err)
		return "", time.Time{}, nil, apperrors.ErrGoogleExchangeFailed
	}

	if !identity.EmailVerified {
		return "", time.Time{}, nil, apperrors.ErrGoogleEmailUnverified
	}

	user, err := s.resolveGoogleUser(ctx, identity)
	if err != nil {
		return "", time.Time{}, nil, err
	}

	token, expiresAt, err := s.tokens.Generate(user.ID)
	if err != nil {
		return "", time.Time{}, nil, err
	}

	return token, expiresAt, user, nil
}

// resolveGoogleUser retrouve le compte lié à une identité Google, le rattache ou le crée
func (s *AuthService) resolveGoogleUser(ctx context.Context, identity *auth.GoogleIdentity) (*models.User, error) {
	user, err := s.users.GetByGoogleID(ctx, identity.Subject)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	email := normalizeEmail(identity.Email)

	user, err = s.users.GetByEmail(ctx, email)
	if err == nil {
		return s.users.LinkGoogle(ctx, user.ID, identity.Subject)
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	return s.users.CreateWithGoogle(ctx, email, googleDisplayName(identity), identity.Subject)
}

// googleDisplayName retient le nom fourni par Google, ou la partie locale de l'email s'il est absent
func googleDisplayName(identity *auth.GoogleIdentity) string {
	name := strings.TrimSpace(identity.Name)
	if name != "" {
		return name
	}

	local, _, _ := strings.Cut(identity.Email, "@")
	if local != "" {
		return local
	}

	return "Utilisateur"
}
