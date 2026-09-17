// Package apperrors définit les erreurs métier de l'application.
//
// Le principe est le suivant : les couches basses (repository, service) ne
// connaissent pas le protocole HTTP. Elles retournent ces erreurs métier, et
// une seule fonction, côté transport HTTP, se charge de les traduire en code
// de statut. Cela évite de disperser des http.Error un peu partout et garantit
// des réponses cohérentes sur toute l'API.
package apperrors

import "errors"

// Les erreurs sont regroupées par le statut HTTP en lequel respondError les
// traduit. Ce classement rend visible, à la lecture, ce que l'API répondra
// dans chaque cas — et fait apparaître immédiatement l'endroit où ranger une
// nouvelle erreur.
var (
	// --- Traduites en 404 Not Found ---

	// ErrNotFound signale qu'une ressource demandée n'existe pas, ou qu'elle
	// n'appartient pas à l'utilisateur qui la demande. Les deux cas sont
	// volontairement confondus : répondre différemment permettrait de deviner
	// l'existence des ressources des autres utilisateurs.
	ErrNotFound = errors.New("ressource introuvable")

	// --- Traduites en 401 Unauthorized ---

	// ErrInvalidCredentials signale un échec de connexion. Le message reste
	// volontairement vague pour ne pas indiquer si c'est l'email ou le mot de
	// passe qui est erroné.
	ErrInvalidCredentials = errors.New("email ou mot de passe incorrect")

	// ErrUnauthorized signale une requête sans jeton d'authentification valide
	// sur une ressource protégée.
	ErrUnauthorized = errors.New("authentification requise")

	// --- Traduites en 409 Conflict ---

	// ErrEmailAlreadyUsed signale une tentative d'inscription avec une adresse
	// email déjà enregistrée.
	ErrEmailAlreadyUsed = errors.New("cette adresse email est déjà utilisée")

	// ErrSpaceNameAlreadyUsed signale qu'un espace du même nom existe déjà chez
	// cet utilisateur. La contrainte est propre à chaque utilisateur : deux
	// personnes différentes peuvent toutes deux avoir un espace « Devoirs ».
	ErrSpaceNameAlreadyUsed = errors.New("un espace portant ce nom existe déjà")

	// --- Traduites en 400 Bad Request ---

	// ErrInvalidNoteStatus signale un état de note hors des valeurs autorisées.
	// La validation des tags `binding` couvre déjà ce cas pour les requêtes de
	// l'API ; cette erreur protège le service s'il est appelé depuis un autre
	// point d'entrée.
	ErrInvalidNoteStatus = errors.New("état de note invalide : valeurs autorisées todo, in_progress, done")
)
