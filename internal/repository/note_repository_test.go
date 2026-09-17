package repository

import (
	"errors"
	"testing"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
	"github.com/KxroTM/test-technique-ynov/internal/models"
)

func TestNoteRepositoryCreateAndGet(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewNoteRepository(db)
	userID := createTestUser(t, db)
	spaceID := createTestSpace(t, db, userID, "Devoirs")

	created, err := repo.Create(ctx, userID, spaceID, "Réviser SQL", "Jointures et index", models.StatusTodo)
	if err != nil {
		t.Fatalf("Create a échoué : %v", err)
	}

	if created.SpaceID != spaceID {
		t.Errorf("space_id = %d, attendu %d", created.SpaceID, spaceID)
	}
	if created.Status != models.StatusTodo {
		t.Errorf("status = %q, attendu %q", created.Status, models.StatusTodo)
	}

	fetched, err := repo.GetByID(ctx, userID, created.ID)
	if err != nil {
		t.Fatalf("GetByID a échoué : %v", err)
	}
	if fetched.Title != "Réviser SQL" {
		t.Errorf("title = %q, attendu %q", fetched.Title, "Réviser SQL")
	}
}

// Une note ne peut pas être créée dans un espace inexistant. C'est la
// traduction de la règle « une note appartient obligatoirement à un espace ».
func TestNoteRepositoryRejectsCreationInMissingSpace(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewNoteRepository(db)
	userID := createTestUser(t, db)

	const missingSpaceID = int64(999999999)

	_, err := repo.Create(ctx, userID, missingSpaceID, "Orpheline", "", models.StatusTodo)
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("erreur = %v, attendu ErrNotFound", err)
	}
}

// Test central du contrôle d'accès sur les notes.
//
// Une note ne porte pas de colonne user_id : son propriétaire est celui de son
// espace. Ce test vérifie que la jointure présente dans chaque requête suffit
// à empêcher tout accès croisé.
func TestNoteRepositoryIsolatesUsers(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewNoteRepository(db)
	ownerID := createTestUser(t, db)
	intruderID := createTestUser(t, db)

	ownerSpaceID := createTestSpace(t, db, ownerID, "Espace du propriétaire")

	note, err := repo.Create(ctx, ownerID, ownerSpaceID, "Note privée", "Contenu confidentiel", models.StatusInProgress)
	if err != nil {
		t.Fatalf("Create a échoué : %v", err)
	}

	t.Run("lecture refusée", func(t *testing.T) {
		_, err := repo.GetByID(ctx, intruderID, note.ID)
		if !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("erreur = %v, attendu ErrNotFound", err)
		}
	})

	t.Run("modification refusée", func(t *testing.T) {
		_, err := repo.Update(ctx, intruderID, note.ID, "Détournée", "", models.StatusDone)
		if !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("erreur = %v, attendu ErrNotFound", err)
		}
	})

	t.Run("suppression refusée", func(t *testing.T) {
		err := repo.Delete(ctx, intruderID, note.ID)
		if !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("erreur = %v, attendu ErrNotFound", err)
		}
	})

	t.Run("listing de l'espace d'autrui vide", func(t *testing.T) {
		notes, err := repo.ListBySpace(ctx, intruderID, ownerSpaceID)
		if err != nil {
			t.Fatalf("ListBySpace a échoué : %v", err)
		}
		if len(notes) != 0 {
			t.Errorf("%d note(s) visible(s) par l'intrus, attendu 0", len(notes))
		}
	})

	// Création dans l'espace d'autrui : le refus doit se traduire par une
	// absence d'insertion, et pas seulement par un code d'erreur.
	t.Run("création dans l'espace d'autrui sans insertion", func(t *testing.T) {
		const intrusionTitle = "Note insérée par un intrus"

		_, err := repo.Create(ctx, intruderID, ownerSpaceID, intrusionTitle, "", models.StatusTodo)
		if !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("erreur = %v, attendu ErrNotFound", err)
		}

		if count := countNotesWithTitle(t, db, intrusionTitle); count != 0 {
			t.Errorf("%d note(s) insérée(s) malgré le refus, attendu 0", count)
		}
	})

	t.Run("note intacte pour le propriétaire", func(t *testing.T) {
		fetched, err := repo.GetByID(ctx, ownerID, note.ID)
		if err != nil {
			t.Fatalf("GetByID pour le propriétaire a échoué : %v", err)
		}
		if fetched.Title != "Note privée" || fetched.Status != models.StatusInProgress {
			t.Errorf("note modifiée : title = %q, status = %q", fetched.Title, fetched.Status)
		}
	})
}

func TestNoteRepositoryUpdate(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewNoteRepository(db)
	userID := createTestUser(t, db)
	spaceID := createTestSpace(t, db, userID, "Jobs")

	note, err := repo.Create(ctx, userID, spaceID, "Avant", "Contenu initial", models.StatusTodo)
	if err != nil {
		t.Fatalf("Create a échoué : %v", err)
	}

	updated, err := repo.Update(ctx, userID, note.ID, "Après", "Contenu révisé", models.StatusDone)
	if err != nil {
		t.Fatalf("Update a échoué : %v", err)
	}

	if updated.Title != "Après" {
		t.Errorf("title = %q, attendu %q", updated.Title, "Après")
	}
	if updated.Status != models.StatusDone {
		t.Errorf("status = %q, attendu %q", updated.Status, models.StatusDone)
	}
	if !updated.UpdatedAt.After(note.UpdatedAt) {
		t.Error("updated_at n'a pas été actualisé")
	}

	// La note ne doit pas avoir changé d'espace : Update ne touche pas à
	// space_id.
	if updated.SpaceID != spaceID {
		t.Errorf("space_id = %d, attendu %d : la note a changé d'espace", updated.SpaceID, spaceID)
	}
}

func TestNoteRepositoryDelete(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewNoteRepository(db)
	userID := createTestUser(t, db)
	spaceID := createTestSpace(t, db, userID, "Personnel")

	note, err := repo.Create(ctx, userID, spaceID, "À supprimer", "", models.StatusTodo)
	if err != nil {
		t.Fatalf("Create a échoué : %v", err)
	}

	if err := repo.Delete(ctx, userID, note.ID); err != nil {
		t.Fatalf("Delete a échoué : %v", err)
	}

	if _, err := repo.GetByID(ctx, userID, note.ID); !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("la note est encore lisible après suppression (erreur = %v)", err)
	}

	if err := repo.Delete(ctx, userID, note.ID); !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("seconde suppression : erreur = %v, attendu ErrNotFound", err)
	}
}

// Supprimer un espace doit supprimer ses notes, sans laisser de ligne
// orpheline. La cascade est déclarée dans le schéma ; ce test vérifie qu'elle
// s'applique réellement.
func TestNoteRepositoryDeletedWithTheirSpace(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	noteRepo := NewNoteRepository(db)
	spaceRepo := NewSpaceRepository(db)

	userID := createTestUser(t, db)
	spaceID := createTestSpace(t, db, userID, "Espace temporaire")

	note, err := noteRepo.Create(ctx, userID, spaceID, "Note liée", "", models.StatusTodo)
	if err != nil {
		t.Fatalf("Create a échoué : %v", err)
	}

	if err := spaceRepo.Delete(ctx, userID, spaceID); err != nil {
		t.Fatalf("suppression de l'espace a échoué : %v", err)
	}

	if _, err := noteRepo.GetByID(ctx, userID, note.ID); !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("la note a survécu à la suppression de son espace (erreur = %v)", err)
	}
}

// Les notes sont retournées de la plus récente à la plus ancienne : c'est
// l'ordre attendu pour une liste de notes que l'on consulte.
func TestNoteRepositoryListIsOrderedByCreationDesc(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewNoteRepository(db)
	userID := createTestUser(t, db)
	spaceID := createTestSpace(t, db, userID, "Ordre")

	first, err := repo.Create(ctx, userID, spaceID, "Première", "", models.StatusTodo)
	if err != nil {
		t.Fatalf("Create a échoué : %v", err)
	}

	second, err := repo.Create(ctx, userID, spaceID, "Seconde", "", models.StatusTodo)
	if err != nil {
		t.Fatalf("Create a échoué : %v", err)
	}

	notes, err := repo.ListBySpace(ctx, userID, spaceID)
	if err != nil {
		t.Fatalf("ListBySpace a échoué : %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf("%d note(s) retournée(s), attendu 2", len(notes))
	}
	if notes[0].ID != second.ID {
		t.Errorf("première note listée = %d, attendu %d (la plus récente)", notes[0].ID, second.ID)
	}
	if notes[1].ID != first.ID {
		t.Errorf("seconde note listée = %d, attendu %d", notes[1].ID, first.ID)
	}
}
