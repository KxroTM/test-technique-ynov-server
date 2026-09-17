// Package repository contient l'accès aux données.
//
// C'est la seule couche qui écrit du SQL. Elle ne contient aucune règle
// métier : elle traduit des appels de méthodes en requêtes, et les erreurs
// techniques de PostgreSQL en erreurs métier du package apperrors.
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

// codeUniqueViolation est le code d'erreur PostgreSQL renvoyé lorsqu'une
// contrainte d'unicité est violée.
const codeUniqueViolation = "23505"

// UserRepository donne accès à la table users.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository construit le repository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create insère un nouvel utilisateur et retourne l'enregistrement créé.
//
// L'unicité de l'email est vérifiée par la base, pas par une lecture préalable :
// un SELECT suivi d'un INSERT laisserait une fenêtre pendant laquelle deux
// inscriptions simultanées pourraient passer. On tente donc l'insertion et on
// interprète la violation de contrainte.
func (r *UserRepository) Create(ctx context.Context, email, passwordHash, name string) (*models.User, error) {
	const query = `
		INSERT INTO users (email, password_hash, name)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, password_hash, created_at`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email, passwordHash, name).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == codeUniqueViolation {
			return nil, apperrors.ErrEmailAlreadyUsed
		}
		return nil, fmt.Errorf("insertion de l'utilisateur : %w", err)
	}

	return user, nil
}

// GetByEmail retourne l'utilisateur correspondant à l'email fourni.
// Elle retourne apperrors.ErrNotFound si aucun utilisateur ne correspond.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const query = `
		SELECT id, email, name, password_hash, created_at
		FROM users
		WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lecture de l'utilisateur par email : %w", err)
	}

	return user, nil
}

// GetByID retourne l'utilisateur correspondant à l'identifiant fourni.
// Elle retourne apperrors.ErrNotFound si aucun utilisateur ne correspond.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	const query = `
		SELECT id, email, name, password_hash, created_at
		FROM users
		WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lecture de l'utilisateur par identifiant : %w", err)
	}

	return user, nil
}
