package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// ClerkJWTVerifier verifies Clerk-issued JWTs
type ClerkJWTVerifier struct {
	secretKey    string
	issuerURL    string
	jwksCache    *JWKSCache
}

type JWKSCache struct {
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
	ttl       time.Duration
}

func NewClerkJWTVerifier(secretKey, issuerURL string) *ClerkJWTVerifier {
	return &ClerkJWTVerifier{
		secretKey: secretKey,
		issuerURL: issuerURL,
		jwksCache: &JWKSCache{
			keys: make(map[string]*rsa.PublicKey),
			ttl:  1 * time.Hour,
		},
	}
}

// Middleware returns an HTTP middleware that verifies Clerk JWTs
func (v *ClerkJWTVerifier) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			respondError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		tokenString := parts[1]
		userID, err := v.VerifyToken(tokenString)
		if err != nil {
			respondError(w, http.StatusUnauthorized, fmt.Sprintf("invalid token: %v", err))
			return
		}

		// Store user ID in context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// VerifyToken verifies a Clerk JWT and returns the user ID (sub claim)
func (v *ClerkJWTVerifier) VerifyToken(tokenString string) (string, error) {
	// Parse token without verification first to get the key ID
	parser := jwt.NewParser(
		jwt.WithIssuer(v.issuerURL),
		jwt.WithExpirationRequired(),
	)

	token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid in token header")
		}

		key, err := v.getPublicKey(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}

		return key, nil
	})

	if err != nil {
		return "", fmt.Errorf("token verification failed: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", errors.New("missing sub claim")
	}

	return sub, nil
}

func (v *ClerkJWTVerifier) getPublicKey(kid string) (*rsa.PublicKey, error) {
	v.jwksCache.mu.RLock()
	if key, ok := v.jwksCache.keys[kid]; ok && time.Since(v.jwksCache.fetchedAt) < v.jwksCache.ttl {
		v.jwksCache.mu.RUnlock()
		return key, nil
	}
	v.jwksCache.mu.RUnlock()

	// Fetch JWKS
	if err := v.fetchJWKS(); err != nil {
		return nil, err
	}

	v.jwksCache.mu.RLock()
	defer v.jwksCache.mu.RUnlock()

	key, ok := v.jwksCache.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key with kid %s not found", kid)
	}

	return key, nil
}

func (v *ClerkJWTVerifier) fetchJWKS() error {
	jwksURL := v.issuerURL + "/.well-known/jwks.json"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	v.jwksCache.mu.Lock()
	defer v.jwksCache.mu.Unlock()

	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" {
			continue
		}

		nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
		if err != nil {
			continue
		}

		eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
		if err != nil {
			continue
		}

		n := new(big.Int).SetBytes(nBytes)
		e := new(big.Int).SetBytes(eBytes)

		pubKey := &rsa.PublicKey{
			N: n,
			E: int(e.Int64()),
		}

		v.jwksCache.keys[jwk.Kid] = pubKey
	}

	v.jwksCache.fetchedAt = time.Now()
	return nil
}

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return userID
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
