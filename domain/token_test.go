package domain

import (
	"testing"
	"time"
)

func TestNewToken(t *testing.T) {
	expiry := time.Now().Add(time.Hour)
	token := NewToken("access123", "refresh456", expiry)

	if token.AccessToken != "access123" {
		t.Errorf("expected access token 'access123', got '%s'", token.AccessToken)
	}
	if token.RefreshToken != "refresh456" {
		t.Errorf("expected refresh token 'refresh456', got '%s'", token.RefreshToken)
	}
	if !token.Expiry.Equal(expiry) {
		t.Errorf("expected expiry %v, got %v", expiry, token.Expiry)
	}
}

func TestToken_IsValid_WhenNotExpired(t *testing.T) {
	expiry := time.Now().Add(time.Hour)
	token := NewToken("access123", "refresh456", expiry)

	if !token.IsValid() {
		t.Error("token should be valid when not expired")
	}
}

func TestToken_IsValid_WhenExpired(t *testing.T) {
	expiry := time.Now().Add(-time.Hour)
	token := NewToken("access123", "refresh456", expiry)

	if token.IsValid() {
		t.Error("token should be invalid when expired")
	}
}

func TestToken_IsValid_WhenExpiringSoon(t *testing.T) {
	// 5分以内に期限切れになるトークンは無効とみなす
	expiry := time.Now().Add(3 * time.Minute)
	token := NewToken("access123", "refresh456", expiry)

	if token.IsValid() {
		t.Error("token should be invalid when expiring within 5 minutes")
	}
}

func TestToken_NeedsRefresh_WhenExpired(t *testing.T) {
	expiry := time.Now().Add(-time.Hour)
	token := NewToken("access123", "refresh456", expiry)

	if !token.NeedsRefresh() {
		t.Error("expired token should need refresh")
	}
}

func TestToken_NeedsRefresh_WhenNoRefreshToken(t *testing.T) {
	expiry := time.Now().Add(-time.Hour)
	token := NewToken("access123", "", expiry)

	if token.NeedsRefresh() {
		t.Error("token without refresh token should not need refresh (cannot refresh)")
	}
}

func TestToken_HasRefreshToken(t *testing.T) {
	token := NewToken("access123", "refresh456", time.Now())
	if !token.HasRefreshToken() {
		t.Error("token should have refresh token")
	}

	tokenNoRefresh := NewToken("access123", "", time.Now())
	if tokenNoRefresh.HasRefreshToken() {
		t.Error("token should not have refresh token")
	}
}

func TestToken_MaskedAccessToken(t *testing.T) {
	token := NewToken("access123456789012345", "refresh456", time.Now())
	masked := token.MaskedAccessToken()

	if masked != "access12345678901234..." {
		t.Errorf("expected 'access12345678901234...', got '%s'", masked)
	}
}

func TestToken_MaskedAccessToken_ShortToken(t *testing.T) {
	token := NewToken("short", "refresh456", time.Now())
	masked := token.MaskedAccessToken()

	if masked != "short" {
		t.Errorf("expected 'short', got '%s'", masked)
	}
}
