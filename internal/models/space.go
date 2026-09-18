package models

import "time"

// Space représente un espace de classement appartenant à un utilisateur
type Space struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	NoteCount int `json:"note_count"`
}
