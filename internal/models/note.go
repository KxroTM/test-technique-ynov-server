package models

import "time"

// NoteStatus représente l'état d'avancement d'une note
type NoteStatus string

const (
	StatusTodo       NoteStatus = "todo"
	StatusInProgress NoteStatus = "in_progress"
	StatusDone       NoteStatus = "done"
)

// IsValid indique si l'état fait partie des valeurs autorisées
func (s NoteStatus) IsValid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// Label retourne le libellé français de l'état, destiné à l'affichage
func (s NoteStatus) Label() string {
	switch s {
	case StatusTodo:
		return "Non fait"
	case StatusInProgress:
		return "En cours"
	case StatusDone:
		return "Terminé"
	default:
		return string(s)
	}
}

// Note représente une note appartenant à un espace
type Note struct {
	ID        int64      `json:"id"`
	SpaceID   int64      `json:"space_id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Status    NoteStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
