package repository

import (
	"errors"
	"testing"

	"github.com/KxroTM/test-technique-ynov/internal/apperrors"
)

func TestSpaceRepositoryCreateAndGet(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewSpaceRepository(db)
	userID := createTestUser(t, db)

	created, err := repo.Create(ctx, userID, "Devoirs", "Travaux scolaires")
	if err != nil {
		t.Fatalf("Create a échoué : %v", err)
	}

	if created.ID == 0 {
		t.Error("l'identifiant de l'espace créé est vide")
	}
	if created.UserID != userID {
		t.Errorf("user_id = %d, attendu %d", created.UserID, userID)
	}
	if created.Name != "Devoirs" {
		t.Errorf("name = %q, attendu %q", created.Name, "Devoirs")
	}

	fetched, err := repo.GetByID(ctx, userID, created.ID)
	if err != nil {
		t.Fatalf("GetByID a échoué : %v", err)
	}
	if fetched.ID != created.ID {
		t.Errorf("identifiant relu = %d, attendu %d", fetched.ID, created.ID)
	}
}

// Deux espaces du même utilisateur ne peuvent pas porter le même nom, quelle
// que soit la casse.
func TestSpaceRepositoryRejectsDuplicateNameForSameUser(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewSpaceRepository(db)
	userID := createTestUser(t, db)

	if _, err := repo.Create(ctx, userID, "Devoirs", ""); err != nil {
		t.Fatalf("Create a échoué : %v", err)
	}

	_, err := repo.Create(ctx, userID, "devoirs", "")
	if !errors.Is(err, apperrors.ErrSpaceNameAlreadyUsed) {
		t.Errorf("erreur = %v, attendu ErrSpaceNameAlreadyUsed", err)
	}
}

// En revanche, deux utilisateurs différents peuvent chacun avoir un espace
// « Devoirs » : la contrainte d'unicité porte sur le couple (utilisateur, nom).
func TestSpaceRepositoryAllowsSameNameForDifferentUsers(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewSpaceRepository(db)
	firstUserID := createTestUser(t, db)
	secondUserID := createTestUser(t, db)

	if _, err := repo.Create(ctx, firstUserID, "Devoirs", ""); err != nil {
		t.Fatalf("Create pour le premier utilisateur a échoué : %v", err)
	}

	if _, err := repo.Create(ctx, secondUserID, "Devoirs", ""); err != nil {
		t.Errorf("Create pour le second utilisateur a échoué : %v", err)
	}
}

// Ce test est le cœur du contrôle d'accès : il vérifie qu'aucune opération
// du repository ne permet d'atteindre l'espace d'un autre utilisateur.
func TestSpaceRepositoryIsolatesUsers(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewSpaceRepository(db)
	ownerID := createTestUser(t, db)
	intruderID := createTestUser(t, db)

	space, err := repo.Create(ctx, ownerID, "Espace privé", "Contenu confidentiel")
	if err != nil {
		t.Fatalf("Create a échoué : %v", err)
	}

	t.Run("lecture refusée", func(t *testing.T) {
		_, err := repo.GetByID(ctx, intruderID, space.ID)
		if !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("erreur = %v, attendu ErrNotFound", err)
		}
	})

	t.Run("modification refusée", func(t *testing.T) {
		_, err := repo.Update(ctx, intruderID, space.ID, "Détourné", "")
		if !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("erreur = %v, attendu ErrNotFound", err)
		}
	})

	t.Run("suppression refusée", func(t *testing.T) {
		err := repo.Delete(ctx, intruderID, space.ID)
		if !errors.Is(err, apperrors.ErrNotFound) {
			t.Errorf("erreur = %v, attendu ErrNotFound", err)
		}
	})

	t.Run("absent de la liste de l'intrus", func(t *testing.T) {
		spaces, err := repo.ListByUser(ctx, intruderID)
		if err != nil {
			t.Fatalf("ListByUser a échoué : %v", err)
		}
		if len(spaces) != 0 {
			t.Errorf("%d espace(s) visible(s) par l'intrus, attendu 0", len(spaces))
		}
	})

	// Après toutes ces tentatives, l'espace doit être intact pour son
	// propriétaire : aucune n'a dû aboutir, même partiellement.
	t.Run("espace intact pour le propriétaire", func(t *testing.T) {
		fetched, err := repo.GetByID(ctx, ownerID, space.ID)
		if err != nil {
			t.Fatalf("GetByID pour le propriétaire a échoué : %v", err)
		}
		if fetched.Name != "Espace privé" {
			t.Errorf("name = %q, attendu %q", fetched.Name, "Espace privé")
		}
	})
}

func TestSpaceRepositoryUpdate(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewSpaceRepository(db)
	userID := createTestUser(t, db)

	space, err := repo.Create(ctx, userID, "Avant", "Description initiale")
	if err != nil {
		t.Fatalf("Create a échoué : %v", err)
	}

	updated, err := repo.Update(ctx, userID, space.ID, "Après", "Nouvelle description")
	if err != nil {
		t.Fatalf("Update a échoué : %v", err)
	}

	if updated.Name != "Après" {
		t.Errorf("name = %q, attendu %q", updated.Name, "Après")
	}
	if updated.Description != "Nouvelle description" {
		t.Errorf("description = %q, attendu %q", updated.Description, "Nouvelle description")
	}
	if !updated.UpdatedAt.After(space.UpdatedAt) {
		t.Error("updated_at n'a pas été actualisé")
	}
	if !updated.CreatedAt.Equal(space.CreatedAt) {
		t.Error("created_at a été modifié alors qu'il doit rester inchangé")
	}
}

func TestSpaceRepositoryDelete(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewSpaceRepository(db)
	userID := createTestUser(t, db)

	space, err := repo.Create(ctx, userID, "À supprimer", "")
	if err != nil {
		t.Fatalf("Create a échoué : %v", err)
	}

	if err := repo.Delete(ctx, userID, space.ID); err != nil {
		t.Fatalf("Delete a échoué : %v", err)
	}

	if _, err := repo.GetByID(ctx, userID, space.ID); !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("l'espace est encore lisible après suppression (erreur = %v)", err)
	}

	// Supprimer deux fois le même espace doit être signalé, et non ignoré
	// silencieusement : sinon le client croirait avoir supprimé une
	// ressource qui n'existait pas.
	if err := repo.Delete(ctx, userID, space.ID); !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("seconde suppression : erreur = %v, attendu ErrNotFound", err)
	}
}

func TestSpaceRepositoryListReturnsEmptySliceNotNil(t *testing.T) {
	db := testDB(t)
	ctx := testContext(t)

	repo := NewSpaceRepository(db)
	userID := createTestUser(t, db)

	spaces, err := repo.ListByUser(ctx, userID)
	if err != nil {
		t.Fatalf("ListByUser a échoué : %v", err)
	}

	// L'API doit répondre `[]` et non `null` quand l'utilisateur n'a aucun
	// espace : un slice nil serait sérialisé en null par encoding/json.
	if spaces == nil {
		t.Error("ListByUser a retourné nil, un slice vide est attendu")
	}
	if len(spaces) != 0 {
		t.Errorf("%d espace(s) retourné(s), attendu 0", len(spaces))
	}
}
