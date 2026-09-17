package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KxroTM/test-technique-ynov/internal/models"
	"github.com/KxroTM/test-technique-ynov/internal/service"
)

// NoteHandler expose les endpoints de gestion des notes.
type NoteHandler struct {
	notes *service.NoteService
}

// NewNoteHandler construit le handler.
func NewNoteHandler(noteService *service.NoteService) *NoteHandler {
	return &NoteHandler{notes: noteService}
}

// createNoteRequest décrit le corps attendu à la création d'une note.
//
// L'espace de destination n'apparaît pas ici : il est lu dans l'URL
// (`/spaces/:spaceID/notes`). Le rattachement à l'espace est ainsi porté par
// la route elle-même, et non par une donnée que le client pourrait choisir
// librement dans le corps de la requête.
//
// Le statut est facultatif — `omitempty` autorise son absence — et vaut
// « todo » par défaut. Lorsqu'il est fourni, `oneof` le contraint aux trois
// valeurs autorisées.
type createNoteRequest struct {
	Title   string            `json:"title"   binding:"required,min=1,max=200"`
	Content string            `json:"content" binding:"max=10000"`
	Status  models.NoteStatus `json:"status"  binding:"omitempty,oneof=todo in_progress done"`
}

// updateNoteRequest décrit le corps attendu à la modification d'une note.
//
// Contrairement à la création, le statut est obligatoire : une modification
// remplace l'intégralité de la note, et omettre le statut reviendrait à le
// réinitialiser silencieusement à « todo ».
type updateNoteRequest struct {
	Title   string            `json:"title"   binding:"required,min=1,max=200"`
	Content string            `json:"content" binding:"max=10000"`
	Status  models.NoteStatus `json:"status"  binding:"required,oneof=todo in_progress done"`
}

// spaceWithNotesResponse est la réponse de la consultation d'un espace.
//
// L'espace et ses notes sont retournés ensemble : le client affiche une page
// par espace, et a besoin des deux pour la construire. Les renvoyer dans une
// seule réponse lui évite un second appel.
type spaceWithNotesResponse struct {
	Space *models.Space `json:"space"`
	Notes []models.Note `json:"notes"`
}

// ListBySpace traite GET /api/spaces/:spaceID/notes.
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

// Create traite POST /api/spaces/:spaceID/notes.
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

// Get traite GET /api/notes/:noteID.
//
// Les notes sont adressées directement par leur identifiant, sans passer par
// celui de leur espace : l'identifiant d'une note suffit à la désigner, et
// l'appartenance est de toute façon vérifiée par la jointure SQL. Imposer
// `/spaces/:spaceID/notes/:noteID` ajouterait un paramètre redondant, et un
// second cas d'erreur (note existante mais dans un autre espace).
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

// Update traite PUT /api/notes/:noteID.
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

// Delete traite DELETE /api/notes/:noteID.
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
