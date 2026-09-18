// Package repository contient l'accès aux données

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
	"github.com/KxroTM/test-technique-ynov/internal/models"
)

const codeUniqueViolation = "23505"

// colonnesUtilisateur liste les colonnes lues par toutes les requêtes de ce repository
const colonnesUtilisateur = `id, email, name, password_hash, google_id, created_at`

// UserRepository donne accès à la table users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository construit le repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create insère un nouvel utilisateur et retourne l'enregistrement créé
func (r *UserRepository) Create(ctx context.Context, email, passwordHash, name string) (*models.User, error) {
	const query = `
		INSERT INTO users (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING ` + colonnesUtilisateur

	return r.scanOne(r.db.QueryRowContext(ctx, query, email, passwordHash, name), "insertion de l'utilisateur")
}

// CreateWithGoogle insère un utilisateur authentifié par Google, sans mot de passe
func (r *UserRepository) CreateWithGoogle(ctx context.Context, email, name, googleID string) (*models.User, error) {
	const query = `
		INSERT INTO users (email, password_hash, name, google_id)
		VALUES ($1, NULL, $2, $3)
		RETURNING ` + colonnesUtilisateur

	return r.scanOne(r.db.QueryRowContext(ctx, query, email, name, googleID), "insertion de l'utilisateur Google")
}

// LinkGoogle rattache un compte Google à un utilisateur existant
func (r *UserRepository) LinkGoogle(ctx context.Context, userID int64, googleID string) (*models.User, error) {
	const query = `
		UPDATE users
		SET google_id = $1
		WHERE id = $2
		RETURNING ` + colonnesUtilisateur

	return r.scanOne(r.db.QueryRowContext(ctx, query, googleID, userID), "liaison du compte Google")
}

// GetByEmail retourne l'utilisateur correspondant à l'email fourni
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `SELECT ` + colonnesUtilisateur + ` FROM users WHERE email = $1`

	return r.scanOne(r.db.QueryRowContext(ctx, query, email), "lecture de l'utilisateur par email")
}

// GetByID retourne l'utilisateur correspondant à l'identifiant fourni
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	const query = `SELECT ` + colonnesUtilisateur + ` FROM users WHERE id = $1`

	return r.scanOne(r.db.QueryRowContext(ctx, query, id), "lecture de l'utilisateur par identifiant")
}

// GetByGoogleID retourne l'utilisateur rattaché à un compte Google
func (r *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*models.User, error) {
	const query = `SELECT ` + colonnesUtilisateur + ` FROM users WHERE google_id = $1`

	return r.scanOne(r.db.QueryRowContext(ctx, query, googleID), "lecture de l'utilisateur par compte Google")
}

// scanOne lit une ligne utilisateur et traduit les erreurs techniques en erreurs métier
func (r *UserRepository) scanOne(row *sql.Row, action string) (*models.User, error) {
	var (
		user         models.User
		passwordHash sql.NullString
		googleID     sql.NullString
	)

	err := row.Scan(&user.ID, &user.Email, &user.Name, &passwordHash, &googleID, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == codeUniqueViolation {
			return nil, apperrors.ErrEmailAlreadyUsed
		}
		return nil, fmt.Errorf("%s : %w", action, err)
	}

	user.PasswordHash = passwordHash.String
	user.GoogleID = googleID.String

	return &user, nil
}
