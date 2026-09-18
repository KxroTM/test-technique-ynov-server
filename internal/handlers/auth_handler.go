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

// AuthHandler expose les endpoints d'inscription, de connexion et de profil
type AuthHandler struct {
	auth *service.AuthService
}

// NewAuthHandler construit le handler.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: authService}
}

// registerRequest décrit le corps attendu par l'endpoint d'inscription
type registerRequest struct {
	Email    string `json:"email"    binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Name     string `json:"name"     binding:"required,min=2,max=100"`
}

// loginRequest décrit le corps attendu par l'endpoint de connexion
type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// loginResponse est la réponse renvoyée après une connexion réussie
type loginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      *models.User `json:"user"`
}

// Register traite POST /api/auth/register
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

// Login traite POST /api/auth/login
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

// Me traite GET /api/me et retourne le profil de l'utilisateur authentifié
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

// googleLoginRequest décrit le corps attendu par l'endpoint de connexion Google
type googleLoginRequest struct {
	Code        string `json:"code"         binding:"required"`
	RedirectURI string `json:"redirect_uri" binding:"required,url"`
}

// LoginWithGoogle traite POST /api/auth/google
func (h *AuthHandler) LoginWithGoogle(c *gin.Context) {
	var req googleLoginRequest
	if !bindJSON(c, &req) {
		return
	}

	token, expiresAt, user, err := h.auth.LoginWithGoogle(c.Request.Context(), req.Code, req.RedirectURI)
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
