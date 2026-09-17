package auth

import "testing"

func TestHashPasswordProducesVerifiableHash(t *testing.T) {
	const plainPassword = "motDePasseSolide123"

	hash, err := HashPassword(plainPassword)
	if err != nil {
		t.Fatalf("HashPassword a échoué : %v", err)
	}

	if hash == plainPassword {
		t.Fatal("le mot de passe a été stocké en clair")
	}

	if !CheckPassword(hash, plainPassword) {
		t.Error("le mot de passe correct a été refusé")
	}

	if CheckPassword(hash, "mauvaisMotDePasse") {
		t.Error("un mot de passe incorrect a été accepté")
	}
}

// Le sel aléatoire de bcrypt garantit que deux empreintes du même mot de passe
// diffèrent. Sans cette propriété, deux utilisateurs ayant choisi le même mot
// de passe seraient identifiables dans la base.
func TestHashPasswordUsesRandomSalt(t *testing.T) {
	const plainPassword = "memeMotDePasse123"

	first, err := HashPassword(plainPassword)
	if err != nil {
		t.Fatalf("HashPassword a échoué : %v", err)
	}

	second, err := HashPassword(plainPassword)
	if err != nil {
		t.Fatalf("HashPassword a échoué : %v", err)
	}

	if first == second {
		t.Error("deux empreintes du même mot de passe sont identiques : le sel n'est pas aléatoire")
	}

	// Les deux empreintes doivent malgré tout valider le même mot de passe.
	if !CheckPassword(first, plainPassword) || !CheckPassword(second, plainPassword) {
		t.Error("une des deux empreintes ne valide pas le mot de passe d'origine")
	}
}

func TestCheckPasswordRejectsMalformedHash(t *testing.T) {
	if CheckPassword("ceci-n-est-pas-une-empreinte-bcrypt", "peu importe") {
		t.Error("une empreinte malformée a été acceptée")
	}
}
