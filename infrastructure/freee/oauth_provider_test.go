package freee

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"freee-oauth-app/domain"
)

func TestFreeeOAuthProvider_AuthorizationURL(t *testing.T) {
	provider := NewFreeeOAuthProvider("client_id", "client_secret", "http://localhost/callback")

	url := provider.AuthorizationURL("test_state")

	if !strings.Contains(url, "accounts.secure.freee.co.jp") {
		t.Error("expected freee authorization URL")
	}
	if !strings.Contains(url, "client_id=client_id") {
		t.Error("expected client_id in URL")
	}
	if !strings.Contains(url, "state=test_state") {
		t.Error("expected state in URL")
	}
	if !strings.Contains(url, "redirect_uri=") {
		t.Error("expected redirect_uri in URL")
	}
	if !strings.Contains(url, "response_type=code") {
		t.Error("expected response_type=code in URL")
	}
}

func TestFreeeOAuthProvider_Exchange_Success(t *testing.T) {
	// モックサーバーを作成
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "test_access_token",
			"refresh_token": "test_refresh_token",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
	defer server.Close()

	provider := NewFreeeOAuthProviderWithEndpoint(
		"client_id",
		"client_secret",
		"http://localhost/callback",
		"http://example.com/auth",
		server.URL,
	)

	ctx := context.Background()
	token, err := provider.Exchange(ctx, "auth_code")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "test_access_token" {
		t.Errorf("expected test_access_token, got %s", token.AccessToken)
	}
	if token.RefreshToken != "test_refresh_token" {
		t.Errorf("expected test_refresh_token, got %s", token.RefreshToken)
	}
}

func TestFreeeOAuthProvider_Exchange_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":             "invalid_grant",
			"error_description": "Invalid authorization code",
		})
	}))
	defer server.Close()

	provider := NewFreeeOAuthProviderWithEndpoint(
		"client_id",
		"client_secret",
		"http://localhost/callback",
		"http://example.com/auth",
		server.URL,
	)

	ctx := context.Background()
	_, err := provider.Exchange(ctx, "invalid_code")

	if err == nil {
		t.Error("expected error for invalid code")
	}
}

func TestFreeeOAuthProvider_Refresh_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "new_access_token",
			"refresh_token": "new_refresh_token",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
	defer server.Close()

	provider := NewFreeeOAuthProviderWithEndpoint(
		"client_id",
		"client_secret",
		"http://localhost/callback",
		"http://example.com/auth",
		server.URL,
	)

	oldToken := domain.NewToken("old_access", "old_refresh", time.Now().Add(-time.Hour))
	ctx := context.Background()
	newToken, err := provider.Refresh(ctx, oldToken)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newToken.AccessToken != "new_access_token" {
		t.Errorf("expected new_access_token, got %s", newToken.AccessToken)
	}
}
