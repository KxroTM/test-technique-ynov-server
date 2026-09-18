package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// adresse à laquelle un code d'autorisation s'échange contre un jeton d'identité
const googleTokenEndpoint = "https://oauth2.googleapis.com/token"

// informations d'un compte Google après échange du code
type GoogleIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
}

// GoogleExchanger échange un code d'autorisation Google contre l'identité du titulaire du compte
type GoogleExchanger struct {
	clientID     string
	clientSecret string
	http         *http.Client
	endpoint     string
}

// NewGoogleExchanger construit l'échangeur à partir des identifiants OAuth de l'application
func NewGoogleExchanger(clientID, clientSecret string) *GoogleExchanger {
	return &GoogleExchanger{
		clientID:     clientID,
		clientSecret: clientSecret,
		http:         &http.Client{Timeout: 10 * time.Second},
		endpoint:     googleTokenEndpoint,
	}
}

// Exchange convertit un code d'autorisation en identité vérifiée
func (g *GoogleExchanger) Exchange(ctx context.Context, code, redirectURI string) (*GoogleIdentity, error) {
	form := url.Values{
		"code":          {code},
		"client_id":     {g.clientID},
		"client_secret": {g.clientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, g.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("construction de la requête Google : %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := g.http.Do(request)
	if err != nil {
		return nil, fmt.Errorf("appel à Google : %w", err)
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("lecture de la réponse Google : %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google a refusé le code (statut %d)", response.StatusCode)
	}

	var body struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, fmt.Errorf("décodage de la réponse Google : %w", err)
	}
	if body.IDToken == "" {
		return nil, fmt.Errorf("google n'a pas retourné de jeton d'identité")
	}

	return parseGoogleIDToken(body.IDToken)
}

// parseGoogleIDToken lit la charge utile d'un jeton d'identité Google sans vérifier sa signature
func parseGoogleIDToken(idToken string) (*GoogleIdentity, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("jeton d'identité Google malformé")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("charge utile du jeton Google illisible : %w", err)
	}

	var claims struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("décodage des informations du compte Google : %w", err)
	}

	if claims.Subject == "" || claims.Email == "" {
		return nil, fmt.Errorf("google n'a pas fourni d'identifiant ou d'email")
	}

	return &GoogleIdentity{
		Subject:       claims.Subject,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
	}, nil
}
