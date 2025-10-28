package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/candlekeep/zot-artifact-store/internal/models"
	"github.com/golang-jwt/jwt/v4"
)

// JWTValidator validates JWT tokens from Keycloak
type JWTValidator struct {
	keycloakURL string
	realm       string
	publicKeys  map[string]*rsa.PublicKey
	httpClient  *http.Client
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(keycloakURL, realm string) *JWTValidator {
	return &JWTValidator{
		keycloakURL: keycloakURL,
		realm:       realm,
		publicKeys:  make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// KeycloakClaims represents JWT claims from Keycloak
type KeycloakClaims struct {
	jwt.RegisteredClaims
	Email             string                 `json:"email"`
	PreferredUsername string                 `json:"preferred_username"`
	RealmAccess       map[string]interface{} `json:"realm_access"`
	Groups            []string               `json:"groups"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSet represents a set of JSON Web Keys
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// ValidateToken validates a JWT token and extracts user information
func (v *JWTValidator) ValidateToken(tokenString string) (*models.User, error) {
	// Parse token without verification to get the key ID
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &KeycloakClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Get key ID from token header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("token missing kid header")
	}

	// Get public key for verification
	publicKey, err := v.getPublicKey(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Verify and parse token with public key
	claims := &KeycloakClaims{}
	token, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract roles from realm access
	roles := []string{}
	if realmAccess, ok := claims.RealmAccess["roles"].([]interface{}); ok {
		for _, role := range realmAccess {
			if roleStr, ok := role.(string); ok {
				roles = append(roles, roleStr)
			}
		}
	}

	// Build user model
	user := &models.User{
		ID:       claims.Subject,
		Username: claims.PreferredUsername,
		Email:    claims.Email,
		Roles:    roles,
		Groups:   claims.Groups,
	}

	return user, nil
}

// getPublicKey fetches and caches the public key for token verification
func (v *JWTValidator) getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check cache first
	if key, ok := v.publicKeys[kid]; ok {
		return key, nil
	}

	// Fetch JWK set from Keycloak
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", v.keycloakURL, v.realm)
	resp, err := v.httpClient.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Find the key with matching kid
	for _, jwk := range jwks.Keys {
		if jwk.Kid == kid {
			publicKey, err := v.jwkToPublicKey(&jwk)
			if err != nil {
				return nil, fmt.Errorf("failed to convert JWK to public key: %w", err)
			}

			// Cache the key
			v.publicKeys[kid] = publicKey
			return publicKey, nil
		}
	}

	return nil, fmt.Errorf("public key not found for kid: %s", kid)
}

// jwkToPublicKey converts a JWK to an RSA public key
func (v *JWTValidator) jwkToPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	// Decode modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert exponent bytes to int
	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}, nil
}

// ExtractBearerToken extracts the bearer token from the Authorization header
func ExtractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	return parts[1], nil
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserContextKey is the context key for the authenticated user
	UserContextKey contextKey = "user"
	// AuthContextKey is the context key for the auth context
	AuthContextKey contextKey = "authContext"
)

// GetUserFromContext extracts the user from the request context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}

// GetAuthContextFromContext extracts the auth context from the request context
func GetAuthContextFromContext(ctx context.Context) (*models.AuthContext, bool) {
	authCtx, ok := ctx.Value(AuthContextKey).(*models.AuthContext)
	return authCtx, ok
}
