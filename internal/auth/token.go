package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager génère et vérifie les jetons JWT de l'application.
//
// Le secret de signature est injecté à la construction plutôt que lu depuis
// l'environnement ici : le package reste ainsi testable sans dépendre de la
// configuration du processus.
type TokenManager struct {
	secret     []byte
	expiration time.Duration
}

// NewTokenManager construit un gestionnaire de jetons.
func NewTokenManager(secret string, expiration time.Duration) *TokenManager {
	return &TokenManager{
		secret:     []byte(secret),
		expiration: expiration,
	}
}

// Generate produit un jeton signé identifiant l'utilisateur fourni.
//
// L'identifiant de l'utilisateur est placé dans le champ standard `sub`
// (subject) plutôt que dans un champ personnalisé : c'est l'usage prévu par
// la spécification JWT pour désigner le porteur du jeton.
//
// Elle retourne également la date d'expiration, que le client utilise pour
// aligner la durée de vie de son cookie de session sur celle du jeton.
func (m *TokenManager) Generate(userID int64) (string, time.Time, error) {
	expiresAt := time.Now().Add(m.expiration)

	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signature du jeton : %w", err)
	}

	return signed, expiresAt, nil
}

// Parse vérifie la signature et la validité d'un jeton, puis retourne
// l'identifiant de l'utilisateur qu'il désigne.
//
// L'algorithme attendu est imposé explicitement via WithValidMethods. Sans
// cette précaution, un jeton forgé avec l'algorithme "none" ou avec un
// algorithme asymétrique pourrait être accepté : c'est une faille classique
// des implémentations JWT.
func (m *TokenManager) Parse(tokenString string) (int64, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwt.RegisteredClaims{},
		func(t *jwt.Token) (interface{}, error) { return m.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return 0, fmt.Errorf("jeton invalide : %w", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return 0, fmt.Errorf("jeton invalide : format des données inattendu")
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("jeton invalide : identifiant utilisateur illisible")
	}

	return userID, nil
}
