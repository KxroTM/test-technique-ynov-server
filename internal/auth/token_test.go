package auth

import (
	"strings"
	"testing"
	"time"
)

const testSecret = "secret-de-test"

func TestGenerateAndParseRoundTrip(t *testing.T) {
	manager := NewTokenManager(testSecret, time.Hour)

	const expectedUserID = int64(42)

	token, expiresAt, err := manager.Generate(expectedUserID)
	if err != nil {
		t.Fatalf("Generate a échoué : %v", err)
	}

	if !expiresAt.After(time.Now()) {
		t.Error("la date d'expiration retournée est déjà passée")
	}

	userID, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse a échoué sur un jeton valide : %v", err)
	}

	if userID != expectedUserID {
		t.Errorf("identifiant utilisateur = %d, attendu %d", userID, expectedUserID)
	}
}

// Un jeton signé avec un autre secret doit être rejeté : c'est ce qui empêche
// un tiers de fabriquer ses propres jetons.
func TestParseRejectsTokenSignedWithAnotherSecret(t *testing.T) {
	attacker := NewTokenManager("mauvais-secret", time.Hour)
	server := NewTokenManager(testSecret, time.Hour)

	forgedToken, _, err := attacker.Generate(1)
	if err != nil {
		t.Fatalf("Generate a échoué : %v", err)
	}

	if _, err := server.Parse(forgedToken); err == nil {
		t.Error("un jeton signé avec un autre secret a été accepté")
	}
}

// Un jeton expiré doit être refusé même si sa signature est valide.
func TestParseRejectsExpiredToken(t *testing.T) {
	// Durée de vie négative : le jeton est expiré dès sa création.
	manager := NewTokenManager(testSecret, -time.Hour)

	expiredToken, _, err := manager.Generate(1)
	if err != nil {
		t.Fatalf("Generate a échoué : %v", err)
	}

	if _, err := manager.Parse(expiredToken); err == nil {
		t.Error("un jeton expiré a été accepté")
	}
}

// Modifier un seul caractère de la signature doit invalider le jeton.
//
// L'altération porte sur le PREMIER caractère de la signature et non sur le
// dernier. Une signature HS256 fait 32 octets, soit 43 caractères base64url
// qui en encodent 258 bits : les deux derniers bits ne servent à rien.
// Plusieurs caractères finaux différents décodent donc vers les mêmes octets,
// et remplacer le dernier caractère ne modifie pas toujours la signature
// réelle. Le premier caractère, lui, porte des bits significatifs.
func TestParseRejectsTamperedToken(t *testing.T) {
	manager := NewTokenManager(testSecret, time.Hour)

	token, _, err := manager.Generate(1)
	if err != nil {
		t.Fatalf("Generate a échoué : %v", err)
	}

	separator := strings.LastIndex(token, ".")
	if separator == -1 || separator == len(token)-1 {
		t.Fatalf("jeton inattendu, signature absente : %q", token)
	}

	// On remplace le premier caractère de la signature par un autre, choisi
	// pour être systématiquement différent de celui d'origine.
	signatureStart := separator + 1
	replacement := byte('A')
	if token[signatureStart] == replacement {
		replacement = 'B'
	}

	tampered := token[:signatureStart] + string(replacement) + token[signatureStart+1:]
	if tampered == token {
		t.Fatal("le jeton altéré est identique à l'original : le test ne vérifie rien")
	}

	if _, err := manager.Parse(tampered); err == nil {
		t.Error("un jeton dont la signature a été modifiée a été accepté")
	}
}

// Un jeton dont l'en-tête annonce alg=none doit être rejeté.
// C'est la faille classique des implémentations qui font confiance à
// l'algorithme déclaré dans le jeton lui-même.
func TestParseRejectsUnsignedToken(t *testing.T) {
	manager := NewTokenManager(testSecret, time.Hour)

	// {"alg":"none","typ":"JWT"}.{"sub":"1","exp":9999999999}. (sans signature)
	const unsignedToken = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0." +
		"eyJzdWIiOiIxIiwiZXhwIjo5OTk5OTk5OTk5fQ."

	if _, err := manager.Parse(unsignedToken); err == nil {
		t.Error("un jeton non signé (alg=none) a été accepté")
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	manager := NewTokenManager(testSecret, time.Hour)

	for _, invalidToken := range []string{"", "pas-un-jeton", strings.Repeat("a.b.c", 3)} {
		if _, err := manager.Parse(invalidToken); err == nil {
			t.Errorf("le jeton invalide %q a été accepté", invalidToken)
		}
	}
}
