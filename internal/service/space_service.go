package service

import (
	"context"
	"strings"

	"github.com/KxroTM/test-technique-ynov/internal/models"
	"github.com/KxroTM/test-technique-ynov/internal/repository"
)

// SpaceService gère les espaces d'un utilisateur
type SpaceService struct {
	spaces *repository.SpaceRepository
}

// NewSpaceService construit le service
func NewSpaceService(spaces *repository.SpaceRepository) *SpaceService {
	return &SpaceService{spaces: spaces}
}

// List retourne les espaces de l'utilisateur
func (s *SpaceService) List(ctx context.Context, userID int64) ([]models.Space, error) {
	return s.spaces.ListByUser(ctx, userID)
}

// Get retourne un espace de l'utilisateur
func (s *SpaceService) Get(ctx context.Context, userID, spaceID int64) (*models.Space, error) {
	return s.spaces.GetByID(ctx, userID, spaceID)
}

// Create crée un espace pour l'utilisateur
func (s *SpaceService) Create(ctx context.Context, userID int64, name, description string) (*models.Space, error) {
	return s.spaces.Create(ctx, userID, strings.TrimSpace(name), strings.TrimSpace(description))
}

// Update modifie un espace de l'utilisateur
func (s *SpaceService) Update(ctx context.Context, userID, spaceID int64, name, description string) (*models.Space, error) {
	return s.spaces.Update(ctx, userID, spaceID, strings.TrimSpace(name), strings.TrimSpace(description))
}

// Delete supprime un espace de l'utilisateur, ainsi que ses notes
func (s *SpaceService) Delete(ctx context.Context, userID, spaceID int64) error {
	return s.spaces.Delete(ctx, userID, spaceID)
}
