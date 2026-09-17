package models

import "time"

// Space représente un espace de classement appartenant à un utilisateur.
// Un espace regroupe des notes autour d'un même thème : « Devoirs »,
// « Jobs », « Personnel »...
type Space struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// NoteCount est le nombre de notes contenues dans l'espace.
	// Ce champ n'existe pas en base : il est calculé par la requête qui
	// liste les espaces, afin d'éviter une requête supplémentaire par
	// espace lors de l'affichage de la liste.
	NoteCount int `json:"note_count"`
}
