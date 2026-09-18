package service

import (
	"context"
	"strings"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
	"github.com/KxroTM/test-technique-ynov/internal/models"
	"github.com/KxroTM/test-technique-ynov/internal/repository"
)

// NoteService gère les notes contenues dans les espaces d'un utilisateur
type NoteService struct {
	notes  *repository.NoteRepository
	spaces *repository.SpaceRepository
}

// NewNoteService construit le service
func NewNoteService(notes *repository.NoteRepository, spaces *repository.SpaceRepository) *NoteService {
	return &NoteService{notes: notes, spaces: spaces}
}

// ListBySpace retourne l'espace demandé accompagné de ses notes
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

// Get retourne une note de l'utilisateur
func (s *NoteService) Get(ctx context.Context, userID, noteID int64) (*models.Note, error) {
	return s.notes.GetByID(ctx, userID, noteID)
}

// Create ajoute une note dans un espace de l'utilisateur
func (s *NoteService) Create(ctx context.Context, userID, spaceID int64, title, content string, status models.NoteStatus) (*models.Note, error) {
	if status == "" {
		status = models.StatusTodo
	}
	if !status.IsValid() {
		return nil, apperrors.ErrInvalidNoteStatus
	}

	return s.notes.Create(ctx, userID, spaceID, strings.TrimSpace(title), content, status)
}

// Update modifie une note de l'utilisateur
func (s *NoteService) Update(ctx context.Context, userID, noteID int64, title, content string, status models.NoteStatus) (*models.Note, error) {
	if !status.IsValid() {
		return nil, apperrors.ErrInvalidNoteStatus
	}

	return s.notes.Update(ctx, userID, noteID, strings.TrimSpace(title), content, status)
}

// Delete supprime une note de l'utilisateur
func (s *NoteService) Delete(ctx context.Context, userID, noteID int64) error {
	return s.notes.Delete(ctx, userID, noteID)
}
