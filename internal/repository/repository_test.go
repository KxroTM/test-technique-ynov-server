package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Les tests de ce package sont des tests d'intégration : ils s'exécutent
// contre une véritable base PostgreSQL, car c'est justement le comportement
// du SQL que l'on veut vérifier (contrôle d'accès dans les clauses WHERE,
// contraintes d'unicité, suppressions en cascade). Les simuler avec un faux
// repository ne prouverait rien.
//
// Si TEST_DATABASE_URL n'est pas défini, les tests sont ignorés plutôt
// qu'en échec : `go test ./...` reste ainsi exécutable sans base disponible.
//
// Pour les lancer :
//   TEST_DATABASE_URL="postgres://notes_user:notes_password@localhost:5432/notes_db?sslmode=disable" go test ./internal/repository -v

// testDB ouvre une connexion vers la base de test, ou ignore le test.
func testDB(t *testing.T) *sql.DB {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL non défini : test d'intégration ignoré")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("ouverture de la base de test : %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("base de test injoignable : %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

// createTestUser insère un utilisateur jetable et programme sa suppression à
// la fin du test.
//
// La suppression de l'utilisateur suffit à nettoyer ses espaces et ses notes :
// les contraintes ON DELETE CASCADE du schéma s'en chargent. Chaque test part
// ainsi d'un état propre sans toucher aux données de démonstration.
func createTestUser(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	// L'horodatage en nanosecondes garantit un email unique, y compris si
	// plusieurs tests créent des utilisateurs dans la même seconde.
	email := fmt.Sprintf("test-%d@example.test", time.Now().UnixNano())

	var userID int64
	err := db.QueryRow(
		`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id`,
		email, "empreinte-de-test", "Utilisateur de test",
	).Scan(&userID)
	if err != nil {
		t.Fatalf("création de l'utilisateur de test : %v", err)
	}

	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM users WHERE id = $1`, userID); err != nil {
			t.Errorf("nettoyage de l'utilisateur de test %d : %v", userID, err)
		}
	})

	return userID
}

// testContext retourne un contexte borné dans le temps, afin qu'un test ne
// puisse pas rester bloqué indéfiniment sur une requête.
func testContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// createTestSpace insère un espace pour l'utilisateur fourni.
//
// Aucun nettoyage explicite n'est programmé : l'espace disparaît avec son
// utilisateur, supprimé par le Cleanup de createTestUser via la cascade.
func createTestSpace(t *testing.T, db *sql.DB, userID int64, name string) int64 {
	t.Helper()

	var spaceID int64
	err := db.QueryRow(
		`INSERT INTO spaces (user_id, name, description) VALUES ($1, $2, $3) RETURNING id`,
		userID, name, "Espace créé par un test",
	).Scan(&spaceID)
	if err != nil {
		t.Fatalf("création de l'espace de test : %v", err)
	}

	return spaceID
}

// countNotesWithTitle compte les notes portant un titre donné.
//
// Elle sert à vérifier qu'une opération refusée n'a réellement rien inséré,
// plutôt que de se contenter du code d'erreur retourné.
func countNotesWithTitle(t *testing.T, db *sql.DB, title string) int {
	t.Helper()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM notes WHERE title = $1`, title).Scan(&count); err != nil {
		t.Fatalf("comptage des notes : %v", err)
	}

	return count
}
