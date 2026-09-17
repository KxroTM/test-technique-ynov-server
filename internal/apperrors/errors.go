// Package apperrors définit les erreurs métier de l'application.
//
// Le principe est le suivant : les couches basses (repository, service) ne
// connaissent pas le protocole HTTP. Elles retournent ces erreurs métier, et
// une seule fonction, côté transport HTTP, se charge de les traduire en code
// de statut. Cela évite de disperser des http.Error un peu partout et garantit
// des réponses cohérentes sur toute l'API.
package apperrors

import "errors"

var (
	// ErrNotFound signale qu'une ressource demandée n'existe pas, ou
	// qu'elle n'appartient pas à l'utilisateur qui la demande. Les deux cas
	// sont volontairement confondus : répondre différemment permettrait de
	// deviner l'existence des ressources des autres utilisateurs.
	ErrNotFound = errors.New("ressource introuvable")

	// ErrEmailAlreadyUsed signale une tentative d'inscription avec une
	// adresse email déjà enregistrée.
	ErrEmailAlreadyUsed = errors.New("cette adresse email est déjà utilisée")

	// ErrInvalidCredentials signale un échec de connexion. Le message reste
	// volontairement vague pour ne pas indiquer si c'est l'email ou le mot
	// de passe qui est erroné.
	ErrInvalidCredentials = errors.New("email ou mot de passe incorrect")

	// ErrUnauthorized signale une requête sans jeton d'authentification
	// valide sur une ressource protégée.
	ErrUnauthorized = errors.New("authentification requise")
)

// ErrSpaceNameAlreadyUsed signale qu'un espace du même nom existe déjà chez
// cet utilisateur. La contrainte est propre à chaque utilisateur : deux
// personnes différentes peuvent toutes deux avoir un espace « Devoirs ».
var ErrSpaceNameAlreadyUsed = errors.New("un espace portant ce nom existe déjà")

// ErrInvalidNoteStatus signale un état de note hors des valeurs autorisées.
// La validation des tags `binding` couvre déjà ce cas pour les requêtes de
// l'API ; cette erreur protège le service s'il est appelé depuis un autre
// point d'entrée.
var ErrInvalidNoteStatus = errors.New("état de note invalide : valeurs autorisées todo, in_progress, done")
