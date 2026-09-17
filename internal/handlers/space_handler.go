package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KxroTM/test-technique-ynov/internal/service"
)

// SpaceHandler expose les endpoints de gestion des espaces.
type SpaceHandler struct {
	spaces *service.SpaceService
}

// NewSpaceHandler construit le handler.
func NewSpaceHandler(spaceService *service.SpaceService) *SpaceHandler {
	return &SpaceHandler{spaces: spaceService}
}

// spaceRequest décrit le corps attendu à la création et à la modification
// d'un espace.
//
// La description est facultative : un espace peut n'avoir qu'un nom. Le nom,
// lui, est obligatoire puisqu'il sert à identifier l'espace pour l'utilisateur.
type spaceRequest struct {
	Name        string `json:"name"        binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=1000"`
}

// List traite GET /api/spaces.
func (h *SpaceHandler) List(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	spaces, err := h.spaces.List(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, spaces)
}

// Get traite GET /api/spaces/:spaceID.
func (h *SpaceHandler) Get(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	spaceID, ok := pathID(c, "spaceID")
	if !ok {
		return
	}

	space, err := h.spaces.Get(c.Request.Context(), userID, spaceID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, space)
}

// Create traite POST /api/spaces.
func (h *SpaceHandler) Create(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req spaceRequest
	if !bindJSON(c, &req) {
		return
	}

	space, err := h.spaces.Create(c.Request.Context(), userID, req.Name, req.Description)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, space)
}

// Update traite PUT /api/spaces/:spaceID.
func (h *SpaceHandler) Update(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	spaceID, ok := pathID(c, "spaceID")
	if !ok {
		return
	}

	var req spaceRequest
	if !bindJSON(c, &req) {
		return
	}

	space, err := h.spaces.Update(c.Request.Context(), userID, spaceID, req.Name, req.Description)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, space)
}

// Delete traite DELETE /api/spaces/:spaceID.
//
// La réponse est un 204 sans corps : il n'y a plus rien à décrire une fois la
// ressource supprimée.
func (h *SpaceHandler) Delete(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	spaceID, ok := pathID(c, "spaceID")
	if !ok {
		return
	}

	if err := h.spaces.Delete(c.Request.Context(), userID, spaceID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
