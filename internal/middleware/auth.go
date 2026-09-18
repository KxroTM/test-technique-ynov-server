// Package middleware contient les intercepteurs HTTP de l'API

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/KxroTM/test-technique-ynov/internal/auth"
)

// contextKeyUserID est la clé sous laquelle l'identifiant de l'utilisateur authentifié est stocké dans le contexte de la requête
const contextKeyUserID = "userID"

// Authenticate vérifie le jeton JWT présent dans l'en-tête Authorization
func Authenticate(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			abortUnauthorized(c, "en-tête Authorization manquant")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abortUnauthorized(c, "format attendu : Authorization: Bearer <jeton>")
			return
		}

		userID, err := tokens.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			abortUnauthorized(c, "jeton invalide ou expiré")
			return
		}

		c.Set(contextKeyUserID, userID)
		c.Next()
	}
}

// UserIDFrom retourne l'identifiant de l'utilisateur authentifié
func UserIDFrom(c *gin.Context) (int64, bool) {
	value, exists := c.Get(contextKeyUserID)
	if !exists {
		return 0, false
	}

	userID, ok := value.(int64)
	return userID, ok
}

// abortUnauthorized interrompt la chaîne de traitement avec un statut 401
func abortUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": message})
}
