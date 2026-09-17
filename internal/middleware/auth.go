// Package middleware contient les intercepteurs HTTP de l'API.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/KxroTM/test-technique-ynov/internal/auth"
)

// contextKeyUserID est la clé sous laquelle l'identifiant de l'utilisateur
// authentifié est stocké dans le contexte de la requête.
const contextKeyUserID = "userID"

// Authenticate vérifie le jeton JWT présent dans l'en-tête Authorization.
//
// En cas de succès, l'identifiant de l'utilisateur est déposé dans le contexte
// de la requête : les handlers protégés le récupèrent avec UserIDFrom et n'ont
// donc jamais à faire confiance à un identifiant transmis par le client.
//
// C'est le point central du contrôle d'accès : toute route enregistrée derrière
// ce middleware est inaccessible sans jeton valide.
func Authenticate(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			abortUnauthorized(c, "en-tête Authorization manquant")
			return
		}

		// Le format attendu est « Bearer <jeton> ».
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

// UserIDFrom retourne l'identifiant de l'utilisateur authentifié.
//
// Le second retour est false si la requête n'est pas passée par le middleware
// Authenticate. Un handler protégé qui obtient false est le signe d'une route
// mal câblée : il doit refuser la requête plutôt que de continuer.
func UserIDFrom(c *gin.Context) (int64, bool) {
	value, exists := c.Get(contextKeyUserID)
	if !exists {
		return 0, false
	}

	userID, ok := value.(int64)
	return userID, ok
}

// abortUnauthorized interrompt la chaîne de traitement avec un statut 401.
// Abort est indispensable : sans lui, Gin appellerait quand même le handler
// suivant après l'écriture de la réponse.
func abortUnauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": message})
}
