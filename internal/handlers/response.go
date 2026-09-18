// Package handlers contient les handlers HTTP de l'API

package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
)

// ErrorResponse est le format unique des réponses d'erreur de l'API
type ErrorResponse struct {
	Error string `json:"error"`

	Fields map[string]string `json:"fields,omitempty"`
}

// respondError traduit une erreur métier en réponse HTTP
func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperrors.ErrInvalidNoteStatus):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})

	case errors.Is(err, apperrors.ErrNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})

	case errors.Is(err, apperrors.ErrEmailAlreadyUsed),
		errors.Is(err, apperrors.ErrSpaceNameAlreadyUsed):
		c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})

	case errors.Is(err, apperrors.ErrGoogleUnavailable):
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: err.Error()})

	case errors.Is(err, apperrors.ErrInvalidCredentials),
		errors.Is(err, apperrors.ErrGoogleEmailUnverified),
		errors.Is(err, apperrors.ErrGoogleExchangeFailed),
		errors.Is(err, apperrors.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})

	default:
		log.Printf("erreur interne : %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "une erreur interne est survenue",
		})
	}
}

// bindJSON décode le corps JSON de la requête dans dst et le valide
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

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "corps de requête illisible : JSON attendu",
		})
		return false
	}
	return true
}

// translateValidationErrors transforme les erreurs du validateur en messages lisibles, indexés par nom de champ JSON
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
