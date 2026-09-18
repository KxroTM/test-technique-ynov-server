package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
	"github.com/KxroTM/test-technique-ynov/internal/middleware"
)

// authenticatedUserID retourne l'identifiant de l'utilisateur authentifié
func authenticatedUserID(c *gin.Context) (int64, bool) {
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		respondError(c, apperrors.ErrUnauthorized)
		return 0, false
	}
	return userID, true
}

// pathID lit un identifiant numérique depuis un paramètre d'URL
func pathID(c *gin.Context, paramName string) (int64, bool) {
	raw := c.Param(paramName)

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "identifiant invalide dans l'URL : un entier positif est attendu",
		})
		return 0, false
	}

	return id, true
}
