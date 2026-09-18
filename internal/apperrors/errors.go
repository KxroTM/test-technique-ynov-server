// Package apperrors définit les erreurs métier de l'application

package apperrors

import "errors"

var (
	// 404 Not Found
	ErrNotFound = errors.New("ressource introuvable")

	// 401 Unauthorized
	ErrInvalidCredentials = errors.New("email ou mot de passe incorrect")
	ErrUnauthorized       = errors.New("authentification requise")

	// 409 Conflict
	ErrEmailAlreadyUsed     = errors.New("cette adresse email est déjà utilisée")
	ErrSpaceNameAlreadyUsed = errors.New("un espace portant ce nom existe déjà")

	// 400 Bad Request
	ErrInvalidNoteStatus = errors.New("état de note invalide : valeurs autorisées todo, in_progress, done")

	// connexion Google pas configurée sur ce serveur
	ErrGoogleUnavailable = errors.New("la connexion Google n'est pas disponible")

	// compte Google dont l'adresse email n'a pas été vérifiée
	ErrGoogleEmailUnverified = errors.New("votre adresse Google doit être vérifiée pour vous connecter")

	// Google a refusé le code d'autorisation
	ErrGoogleExchangeFailed = errors.New("la connexion Google a échoué, veuillez réessayer")
)
