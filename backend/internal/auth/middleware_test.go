package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserID_EmptyContext(t *testing.T) {
	ctx := context.Background()
	userID := GetUserID(ctx)
	assert.Empty(t, userID)
}

func TestGetUserID_WithValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "user_123")
	userID := GetUserID(ctx)
	assert.Equal(t, "user_123", userID)
}

func TestMiddleware_MissingAuthHeader(t *testing.T) {
	verifier := NewClerkJWTVerifier("test-secret", "https://test.clerk.accounts.dev")

	handler := verifier.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing authorization header")
}

func TestMiddleware_InvalidAuthFormat(t *testing.T) {
	verifier := NewClerkJWTVerifier("test-secret", "https://test.clerk.accounts.dev")

	handler := verifier.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid authorization header format")
}

func TestMiddleware_InvalidToken(t *testing.T) {
	verifier := NewClerkJWTVerifier("test-secret", "https://test.clerk.accounts.dev")

	handler := verifier.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.jwt.token")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid token")
}

func TestMiddleware_NoBearerPrefix(t *testing.T) {
	verifier := NewClerkJWTVerifier("test-secret", "https://test.clerk.accounts.dev")

	handler := verifier.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid authorization header format")
}
