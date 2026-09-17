package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
	"github.com/KxroTM/test-technique-ynov/internal/middleware"
)

// authenticatedUserID retourne l'identifiant de l'utilisateur authentifié.
//
// Elle retourne false après avoir déjà écrit la réponse d'erreur, sur le même
// principe que bindJSON : le handler appelant n'a plus qu'à s'arrêter.
func authenticatedUserID(c *gin.Context) (int64, bool) {
	userID, ok := middleware.UserIDFrom(c)
	if !ok {
		respondError(c, apperrors.ErrUnauthorized)
		return 0, false
	}
	return userID, true
}

// pathID lit un identifiant numérique depuis un paramètre d'URL.
//
// Une valeur non numérique est refusée avec un statut 400 et non 404 : la
// requête est malformée, elle ne désigne aucune ressource. Les identifiants
// négatifs ou nuls sont également refusés, ce qui évite d'envoyer des
// requêtes SQL qui ne peuvent rien retourner.
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
