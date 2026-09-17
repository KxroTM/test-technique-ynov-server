// Package handlers contient les handlers HTTP de l'API.
//
// Leur rôle est volontairement limité : décoder et valider la requête,
// appeler le service concerné, puis sérialiser le résultat. Aucune règle
// métier n'est écrite ici.
package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
)

// ErrorResponse est le format unique des réponses d'erreur de l'API.
// Utiliser la même structure partout permet au client de traiter toutes les
// erreurs avec un seul chemin de code.
type ErrorResponse struct {
	Error string `json:"error"`

	// Fields détaille les erreurs de validation, champ par champ.
	// Il est absent des réponses qui ne concernent pas la validation.
	Fields map[string]string `json:"fields,omitempty"`
}

// respondError traduit une erreur métier en réponse HTTP.
//
// C'est le seul endroit de l'application où une erreur devient un code de
// statut. Les couches service et repository restent ainsi indépendantes du
// protocole HTTP, et les réponses sont cohérentes sur toute l'API.
func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperrors.ErrInvalidNoteStatus):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})

	case errors.Is(err, apperrors.ErrNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})

	case errors.Is(err, apperrors.ErrEmailAlreadyUsed),
		errors.Is(err, apperrors.ErrSpaceNameAlreadyUsed):
		c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})

	case errors.Is(err, apperrors.ErrInvalidCredentials),
		errors.Is(err, apperrors.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})

	default:
		// Erreur inattendue : elle est journalisée côté serveur avec son
		// détail, mais le client ne reçoit qu'un message générique. Renvoyer
		// l'erreur brute pourrait exposer la structure de la base ou des
		// informations d'infrastructure.
		log.Printf("erreur interne : %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "une erreur interne est survenue",
		})
	}
}

// bindJSON décode le corps JSON de la requête dans dst et le valide.
//
// Elle retourne false si la requête est invalide, après avoir déjà écrit la
// réponse d'erreur : le handler appelant n'a plus qu'à s'arrêter.
func bindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:  "données invalides",
				Fields: translateValidationErrors(validationErrors),
			})
			return false
		}

		// Le corps n'est pas un JSON exploitable (syntaxe invalide,
		// type incorrect...).
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "corps de requête illisible : JSON attendu",
		})
		return false
	}
	return true
}

// translateValidationErrors transforme les erreurs du validateur en messages
// lisibles, indexés par nom de champ JSON.
func translateValidationErrors(errs validator.ValidationErrors) map[string]string {
	messages := make(map[string]string, len(errs))

	for _, fieldErr := range errs {
		field := fieldErr.Field()

		switch fieldErr.Tag() {
		case "required":
			messages[field] = "ce champ est obligatoire"
		case "email":
			messages[field] = "adresse email invalide"
		case "min":
			messages[field] = "ce champ doit contenir au moins " + fieldErr.Param() + " caractères"
		case "max":
			messages[field] = "ce champ ne doit pas dépasser " + fieldErr.Param() + " caractères"
		case "oneof":
			messages[field] = "valeur attendue parmi : " + fieldErr.Param()
		default:
			messages[field] = "valeur invalide"
		}
	}

	return messages
}
