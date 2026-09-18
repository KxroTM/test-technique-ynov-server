package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KxroTM/test-technique-ynov/internal/models"
	"github.com/KxroTM/test-technique-ynov/internal/service"
)

// NoteHandler expose les endpoints de gestion des notes
type NoteHandler struct {
	notes *service.NoteService
}

// NewNoteHandler construit le handler
func NewNoteHandler(noteService *service.NoteService) *NoteHandler {
	return &NoteHandler{notes: noteService}
}

// createNoteRequest décrit le corps attendu à la création d'une note
type createNoteRequest struct {
	Title   string            `json:"title"   binding:"required,min=1,max=200"`
	Content string            `json:"content" binding:"max=10000"`
	Status  models.NoteStatus `json:"status"  binding:"omitempty,oneof=todo in_progress done"`
}

// updateNoteRequest décrit le corps attendu à la modification d'une note
type updateNoteRequest struct {
	Title   string            `json:"title"   binding:"required,min=1,max=200"`
	Content string            `json:"content" binding:"max=10000"`
	Status  models.NoteStatus `json:"status"  binding:"required,oneof=todo in_progress done"`
}

// spaceWithNotesResponse est la réponse de la consultation d'un espace
type spaceWithNotesResponse struct {
	Space *models.Space `json:"space"`
	Notes []models.Note `json:"notes"`
}

// ListBySpace traite GET /api/spaces/:spaceID/notes
func (h *NoteHandler) ListBySpace(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	spaceID, ok := pathID(c, "spaceID")
	if !ok {
		return
	}

	space, notes, err := h.notes.ListBySpace(c.Request.Context(), userID, spaceID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, spaceWithNotesResponse{Space: space, Notes: notes})
}

// Create traite POST /api/spaces/:spaceID/notes
func (h *NoteHandler) Create(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	spaceID, ok := pathID(c, "spaceID")
	if !ok {
		return
	}

	var req createNoteRequest
	if !bindJSON(c, &req) {
		return
	}

	note, err := h.notes.Create(c.Request.Context(), userID, spaceID, req.Title, req.Content, req.Status)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, note)
}

// Get traite GET /api/notes/:noteID
func (h *NoteHandler) Get(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	noteID, ok := pathID(c, "noteID")
	if !ok {
		return
	}

	note, err := h.notes.Get(c.Request.Context(), userID, noteID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, note)
}

// Update traite PUT /api/notes/:noteID
func (h *NoteHandler) Update(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	noteID, ok := pathID(c, "noteID")
	if !ok {
		return
	}

	var req updateNoteRequest
	if !bindJSON(c, &req) {
		return
	}

	note, err := h.notes.Update(c.Request.Context(), userID, noteID, req.Title, req.Content, req.Status)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, note)
}

// Delete traite DELETE /api/notes/:noteID
func (h *NoteHandler) Delete(c *gin.Context) {
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	noteID, ok := pathID(c, "noteID")
	if !ok {
		return
	}

	if err := h.notes.Delete(c.Request.Context(), userID, noteID); err != nil {
		respondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
