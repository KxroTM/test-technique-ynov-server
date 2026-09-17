package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
	"github.com/KxroTM/test-technique-ynov/internal/models"
)

// NoteRepository donne accès à la table notes.
//
// Comme pour les espaces, toutes les méthodes exigent l'identifiant de
// l'utilisateur. La difficulté est ici supplémentaire : une note n'a pas de
// colonne user_id. Son propriétaire est l'utilisateur de l'espace qui la
// contient. Chaque requête effectue donc une jointure sur spaces afin de
// vérifier l'appartenance, plutôt que de faire confiance à un contrôle
// réalisé en amont dans le code Go.
type NoteRepository struct {
	db *sql.DB
}

// NewNoteRepository construit le repository.
func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// ListBySpace retourne les notes d'un espace appartenant à l'utilisateur.
//
// La jointure sur spaces assure que la liste est vide si l'espace appartient
// à quelqu'un d'autre : il n'y a donc aucune fuite de contenu, même si
// l'appelant fournit un identifiant d'espace au hasard.
func (r *NoteRepository) ListBySpace(ctx context.Context, userID, spaceID int64) ([]models.Note, error) {
	const query = `
		SELECT n.id, n.space_id, n.title, n.content, n.status, n.created_at, n.updated_at
		FROM notes n
		JOIN spaces s ON s.id = n.space_id
		WHERE n.space_id = $1 AND s.user_id = $2
		ORDER BY n.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, spaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("lecture des notes : %w", err)
	}
	defer rows.Close()

	notes := []models.Note{}

	for rows.Next() {
		var note models.Note
		err := rows.Scan(
			&note.ID,
			&note.SpaceID,
			&note.Title,
			&note.Content,
			&note.Status,
			&note.CreatedAt,
			&note.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("lecture d'une ligne de note : %w", err)
		}
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours des notes : %w", err)
	}

	return notes, nil
}

// GetByID retourne une note appartenant à l'utilisateur fourni.
func (r *NoteRepository) GetByID(ctx context.Context, userID, noteID int64) (*models.Note, error) {
	const query = `
		SELECT n.id, n.space_id, n.title, n.content, n.status, n.created_at, n.updated_at
		FROM notes n
		JOIN spaces s ON s.id = n.space_id
		WHERE n.id = $1 AND s.user_id = $2`

	note := &models.Note{}
	err := r.db.QueryRowContext(ctx, query, noteID, userID).Scan(
		&note.ID,
		&note.SpaceID,
		&note.Title,
		&note.Content,
		&note.Status,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lecture de la note : %w", err)
	}

	return note, nil
}

// Create insère une note dans un espace appartenant à l'utilisateur.
//
// L'insertion utilise `INSERT ... SELECT ... WHERE` plutôt qu'un VALUES
// classique : la ligne n'est créée que si la sous-requête confirme que
// l'espace visé appartient bien à l'utilisateur. Une tentative de création
// dans l'espace d'autrui n'insère donc simplement aucune ligne, et la clause
// RETURNING ne renvoie rien.
//
// L'alternative — lire l'espace puis insérer — aurait nécessité deux requêtes
// et laissé une fenêtre pendant laquelle l'espace peut être supprimé.
func (r *NoteRepository) Create(ctx context.Context, userID, spaceID int64, title, content string, status models.NoteStatus) (*models.Note, error) {
	const query = `
		INSERT INTO notes (space_id, title, content, status)
		SELECT $1, $2, $3, $4
		WHERE EXISTS (
			SELECT 1 FROM spaces WHERE id = $1 AND user_id = $5
		)
		RETURNING id, space_id, title, content, status, created_at, updated_at`

	note := &models.Note{}
	err := r.db.QueryRowContext(ctx, query, spaceID, title, content, status, userID).Scan(
		&note.ID,
		&note.SpaceID,
		&note.Title,
		&note.Content,
		&note.Status,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		// Aucune ligne insérée : l'espace n'existe pas, ou il appartient à
		// un autre utilisateur. Les deux cas sont indistinguables, comme
		// partout ailleurs dans l'application.
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("insertion de la note : %w", err)
	}

	return note, nil
}

// Update modifie une note appartenant à l'utilisateur fourni.
//
// La note n'est pas déplaçable d'un espace à l'autre : space_id n'est pas
// modifiable. Le sujet ne le demande pas, et cela évite d'avoir à vérifier
// l'appartenance de l'espace de destination.
func (r *NoteRepository) Update(ctx context.Context, userID, noteID int64, title, content string, status models.NoteStatus) (*models.Note, error) {
	const query = `
		UPDATE notes
		SET title = $1, content = $2, status = $3, updated_at = NOW()
		WHERE id = $4
		  AND space_id IN (SELECT id FROM spaces WHERE user_id = $5)
		RETURNING id, space_id, title, content, status, created_at, updated_at`

	note := &models.Note{}
	err := r.db.QueryRowContext(ctx, query, title, content, status, noteID, userID).Scan(
		&note.ID,
		&note.SpaceID,
		&note.Title,
		&note.Content,
		&note.Status,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("mise à jour de la note : %w", err)
	}

	return note, nil
}

// Delete supprime une note appartenant à l'utilisateur fourni.
func (r *NoteRepository) Delete(ctx context.Context, userID, noteID int64) error {
	const query = `
		DELETE FROM notes
		WHERE id = $1
		  AND space_id IN (SELECT id FROM spaces WHERE user_id = $2)`

	result, err := r.db.ExecContext(ctx, query, noteID, userID)
	if err != nil {
		return fmt.Errorf("suppression de la note : %w", err)
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
