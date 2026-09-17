package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
	"github.com/KxroTM/test-technique-ynov/internal/middleware"
	"github.com/KxroTM/test-technique-ynov/internal/models"
	"github.com/KxroTM/test-technique-ynov/internal/service"
)

// AuthHandler expose les endpoints d'inscription, de connexion et de profil.
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler construit le handler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: authService}
}

// registerRequest décrit le corps attendu par l'endpoint d'inscription.
//
// Les règles de validation sont déclarées par des tags `binding`. Elles sont
// appliquées avant que le handler ne s'exécute : la logique métier ne reçoit
// donc jamais de données de forme invalide.
//
// La limite de 72 caractères sur le mot de passe n'est pas arbitraire :
// bcrypt ignore silencieusement les octets au-delà du 72e. Refuser
// explicitement les mots de passe plus longs évite qu'un utilisateur croie
// être protégé par une phrase secrète dont seule la première partie compte.
type registerRequest struct {
	Email    string `json:"email"    binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Name     string `json:"name"     binding:"required,min=2,max=100"`
}

// loginRequest décrit le corps attendu par l'endpoint de connexion.
//
// Aucune contrainte de longueur minimale n'est posée sur le mot de passe :
// à la connexion, la seule question est de savoir s'il correspond. Imposer
// une règle de format ici renseignerait sur la politique de mots de passe.
type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// loginResponse est la réponse renvoyée après une connexion réussie.
type loginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      *models.User `json:"user"`
}

// Register traite POST /api/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if !bindJSON(c, &req) {
		return
	}

	user, err := h.auth.Register(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login traite POST /api/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if !bindJSON(c, &req) {
		return
	}

	token, expiresAt, user, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	})
}

// Me traite GET /api/me et retourne le profil de l'utilisateur authentifié.
//
// L'identifiant est lu dans le contexte, alimenté par le middleware
// d'authentification. Il n'est jamais lu depuis l'URL ou le corps de la
// requête : sans quoi n'importe qui pourrait demander le profil d'un autre.
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		respondError(c, apperrors.ErrUnauthorized)
		return
	}

	user, err := h.auth.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}
