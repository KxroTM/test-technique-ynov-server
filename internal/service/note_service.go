package service

import (
	"context"
	"strings"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
	"github.com/KxroTM/test-technique-ynov/internal/models"
	"github.com/KxroTM/test-technique-ynov/internal/repository"
)

// NoteService gère les notes contenues dans les espaces d'un utilisateur.
type NoteService struct {
	notes  *repository.NoteRepository
	spaces *repository.SpaceRepository
}

// NewNoteService construit le service.
//
// Le service dépend aussi du repository des espaces : afficher les notes d'un
// espace suppose de connaître l'espace lui-même (son nom, sa description),
// afin que le client puisse titrer la page sans requête supplémentaire.
func NewNoteService(notes *repository.NoteRepository, spaces *repository.SpaceRepository) *NoteService {
	return &NoteService{notes: notes, spaces: spaces}
}

// ListBySpace retourne l'espace demandé accompagné de ses notes.
//
// L'espace est lu en premier : si l'utilisateur n'y a pas accès, on retourne
// ErrNotFound sans même interroger les notes. Cela distingue proprement
// « espace inaccessible » (404) de « espace accessible mais vide » (liste vide).
func (s *NoteService) ListBySpace(ctx context.Context, userID, spaceID int64) (*models.Space, []models.Note, error) {
	space, err := s.spaces.GetByID(ctx, userID, spaceID)
	if err != nil {
		return nil, nil, err
	}

	notes, err := s.notes.ListBySpace(ctx, userID, spaceID)
	if err != nil {
		return nil, nil, err
	}

	return space, notes, nil
}

// Get retourne une note de l'utilisateur.
func (s *NoteService) Get(ctx context.Context, userID, noteID int64) (*models.Note, error) {
	return s.notes.GetByID(ctx, userID, noteID)
}

// Create ajoute une note dans un espace de l'utilisateur.
//
// L'état est optionnel à la création : une note nouvellement ajoutée est
// « non faite » par défaut, ce qui correspond au cas d'usage courant.
func (s *NoteService) Create(ctx context.Context, userID, spaceID int64, title, content string, status models.NoteStatus) (*models.Note, error) {
	if status == "" {
		status = models.StatusTodo
	}
	if !status.IsValid() {
		return nil, apperrors.ErrInvalidNoteStatus
	}

	return s.notes.Create(ctx, userID, spaceID, strings.TrimSpace(title), content, status)
}

// Update modifie une note de l'utilisateur.
func (s *NoteService) Update(ctx context.Context, userID, noteID int64, title, content string, status models.NoteStatus) (*models.Note, error) {
	if !status.IsValid() {
		return nil, apperrors.ErrInvalidNoteStatus
	}

	return s.notes.Update(ctx, userID, noteID, strings.TrimSpace(title), content, status)
}

// Delete supprime une note de l'utilisateur.
func (s *NoteService) Delete(ctx context.Context, userID, noteID int64) error {
	return s.notes.Delete(ctx, userID, noteID)
}
