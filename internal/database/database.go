// Package database gère l'ouverture et la configuration de la connexion
// à PostgreSQL.
package database

import (
	"database/sql"
	"fmt"
	"time"

	// Le driver pgx est enregistré auprès de database/sql par cet import.
	// On garde ainsi l'API standard de la bibliothèque Go plutôt que de
	// dépendre de l'API spécifique du driver.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Connect ouvre une connexion à la base et vérifie qu'elle répond.
//
// sql.Open n'établit pas réellement de connexion : il se contente de préparer
// le pool. On appelle donc Ping pour échouer immédiatement au démarrage si la
// base est injoignable, plutôt qu'à la première requête d'un utilisateur.
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("ouverture de la connexion : %w", err)
	}

	// Limites du pool de connexions. Sans ces réglages, Go autorise un nombre
	// illimité de connexions ouvertes, ce qui peut saturer PostgreSQL.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("base de données injoignable : %w", err)
	}

	return db, nil
}
