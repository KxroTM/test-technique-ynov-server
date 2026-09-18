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

// SpaceRepository donne accès à la table spaces
type SpaceRepository struct {
	db *sql.DB
}

// NewSpaceRepository construit le repository
func NewSpaceRepository(db *sql.DB) *SpaceRepository {
	return &SpaceRepository{db: db}
}

// ListByUser retourne tous les espaces d'un utilisateur, avec le nombre de notes de chacun
func (r *SpaceRepository) ListByUser(ctx context.Context, userID int64) ([]models.Space, error) {
	const query = `
		SELECT s.id, s.user_id, s.name, s.description, s.created_at, s.updated_at,
		       COUNT(n.id) AS note_count
		FROM spaces s
		LEFT JOIN notes n ON n.space_id = s.id
		WHERE s.user_id = $1
		GROUP BY s.id
		ORDER BY s.name`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("lecture des espaces : %w", err)
	}
	defer rows.Close()

	spaces := []models.Space{}

	for rows.Next() {
		var space models.Space
		err := rows.Scan(
			&space.ID,
			&space.UserID,
			&space.Name,
			&space.Description,
			&space.CreatedAt,
			&space.UpdatedAt,
			&space.NoteCount,
		)
		if err != nil {
			return nil, fmt.Errorf("lecture d'une ligne d'espace : %w", err)
		}
		spaces = append(spaces, space)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours des espaces : %w", err)
	}

	return spaces, nil
}

// GetByID retourne un espace appartenant à l'utilisateur fourni
func (r *SpaceRepository) GetByID(ctx context.Context, userID, spaceID int64) (*models.Space, error) {
	const query = `
		SELECT s.id, s.user_id, s.name, s.description, s.created_at, s.updated_at,
		       COUNT(n.id) AS note_count
		FROM spaces s
		LEFT JOIN notes n ON n.space_id = s.id
		WHERE s.id = $1 AND s.user_id = $2
		GROUP BY s.id`

	space := &models.Space{}
	err := r.db.QueryRowContext(ctx, query, spaceID, userID).Scan(
		&space.ID,
		&space.UserID,
		&space.Name,
		&space.Description,
		&space.CreatedAt,
		&space.UpdatedAt,
		&space.NoteCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lecture de l'espace : %w", err)
	}

	return space, nil
}

// Create insère un nouvel espace pour l'utilisateur fourni
func (r *SpaceRepository) Create(ctx context.Context, userID int64, name, description string) (*models.Space, error) {
	const query = `
		INSERT INTO spaces (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, name, description, created_at, updated_at`

	space := &models.Space{}
	err := r.db.QueryRowContext(ctx, query, userID, name, description).Scan(
		&space.ID,
		&space.UserID,
		&space.Name,
		&space.Description,
		&space.CreatedAt,
		&space.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == codeUniqueViolation {
			return nil, apperrors.ErrSpaceNameAlreadyUsed
		}
		return nil, fmt.Errorf("insertion de l'espace : %w", err)
	}

	return space, nil
}

// Update modifie le nom et la description d'un espace appartenant à l'utilisateur fourni
func (r *SpaceRepository) Update(ctx context.Context, userID, spaceID int64, name, description string) (*models.Space, error) {
	const query = `
		UPDATE spaces
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, name, description, created_at, updated_at`

	space := &models.Space{}
	err := r.db.QueryRowContext(ctx, query, name, description, spaceID, userID).Scan(
		&space.ID,
		&space.UserID,
		&space.Name,
		&space.Description,
		&space.CreatedAt,
		&space.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == codeUniqueViolation {
			return nil, apperrors.ErrSpaceNameAlreadyUsed
		}
		return nil, fmt.Errorf("mise à jour de l'espace : %w", err)
	}

	return space, nil
}

// Delete supprime un espace appartenant à l'utilisateur fourni
func (r *SpaceRepository) Delete(ctx context.Context, userID, spaceID int64) error {
	const query = `DELETE FROM spaces WHERE id = $1 AND user_id = $2`

	result, err := r.db.ExecContext(ctx, query, spaceID, userID)
	if err != nil {
		return fmt.Errorf("suppression de l'espace : %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("lecture du nombre de lignes supprimées : %w", err)
	}
	if affected == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}
