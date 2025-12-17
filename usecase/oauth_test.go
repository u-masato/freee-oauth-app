package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"freee-oauth-app/domain"
)

// モックTokenRepository
type mockTokenRepository struct {
	token      *domain.Token
	saveErr    error
	loadErr    error
	saveCalled bool
}

func (m *mockTokenRepository) Save(ctx context.Context, token *domain.Token) error {
	m.saveCalled = true
	if m.saveErr != nil {
		return m.saveErr
	}
	m.token = token
	return nil
}

func (m *mockTokenRepository) Load(ctx context.Context) (*domain.Token, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.token, nil
}

func (m *mockTokenRepository) Exists(ctx context.Context) bool {
	return m.token != nil
}

// モックOAuthProvider
type mockOAuthProvider struct {
	authURL     string
	token       *domain.Token
	exchangeErr error
	refreshErr  error
}

func (m *mockOAuthProvider) AuthorizationURL(state string) string {
	return m.authURL + "?state=" + state
}

func (m *mockOAuthProvider) Exchange(ctx context.Context, code string) (*domain.Token, error) {
	if m.exchangeErr != nil {
		return nil, m.exchangeErr
	}
	return m.token, nil
}

func (m *mockOAuthProvider) Refresh(ctx context.Context, token *domain.Token) (*domain.Token, error) {
	if m.refreshErr != nil {
		return nil, m.refreshErr
	}
	return m.token, nil
}

func TestOAuthUseCase_GetOrRefreshToken_WhenValidTokenExists(t *testing.T) {
	validToken := domain.NewToken("access", "refresh", time.Now().Add(time.Hour))
	repo := &mockTokenRepository{token: validToken}
	provider := &mockOAuthProvider{}
	uc := NewOAuthUseCase(repo, provider)

	token, err := uc.GetOrRefreshToken(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if token != validToken {
		t.Error("expected valid token to be returned")
	}
}

func TestOAuthUseCase_GetOrRefreshToken_WhenNoTokenExists(t *testing.T) {
	repo := &mockTokenRepository{loadErr: errors.New("not found")}
	provider := &mockOAuthProvider{}
	uc := NewOAuthUseCase(repo, provider)

	token, err := uc.GetOrRefreshToken(context.Background())

	if err != ErrNoToken {
		t.Errorf("expected ErrNoToken, got %v", err)
	}
	if token != nil {
		t.Error("expected nil token")
	}
}

func TestOAuthUseCase_GetOrRefreshToken_WhenTokenNeedsRefresh(t *testing.T) {
	expiredToken := domain.NewToken("old_access", "refresh", time.Now().Add(-time.Hour))
	newToken := domain.NewToken("new_access", "refresh", time.Now().Add(time.Hour))
	repo := &mockTokenRepository{token: expiredToken}
	provider := &mockOAuthProvider{token: newToken}
	uc := NewOAuthUseCase(repo, provider)

	token, err := uc.GetOrRefreshToken(context.Background())

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if token.AccessToken != "new_access" {
		t.Errorf("expected new_access, got %s", token.AccessToken)
	}
	if !repo.saveCalled {
		t.Error("expected token to be saved after refresh")
	}
}

func TestOAuthUseCase_GetOrRefreshToken_WhenRefreshFails(t *testing.T) {
	expiredToken := domain.NewToken("old_access", "refresh", time.Now().Add(-time.Hour))
	repo := &mockTokenRepository{token: expiredToken}
	provider := &mockOAuthProvider{refreshErr: errors.New("refresh failed")}
	uc := NewOAuthUseCase(repo, provider)

	token, err := uc.GetOrRefreshToken(context.Background())

	if err != ErrRefreshFailed {
		t.Errorf("expected ErrRefreshFailed, got %v", err)
	}
	if token != nil {
		t.Error("expected nil token")
	}
}

func TestOAuthUseCase_StartAuthorization(t *testing.T) {
	repo := &mockTokenRepository{}
	provider := &mockOAuthProvider{authURL: "https://example.com/auth"}
	uc := NewOAuthUseCase(repo, provider)

	url, state := uc.StartAuthorization()

	if state == "" {
		t.Error("expected non-empty state")
	}
	expectedURL := "https://example.com/auth?state=" + state
	if url != expectedURL {
		t.Errorf("expected %s, got %s", expectedURL, url)
	}
}

func TestOAuthUseCase_CompleteAuthorization_Success(t *testing.T) {
	newToken := domain.NewToken("access", "refresh", time.Now().Add(time.Hour))
	repo := &mockTokenRepository{}
	provider := &mockOAuthProvider{token: newToken}
	uc := NewOAuthUseCase(repo, provider)

	_, state := uc.StartAuthorization()
	token, err := uc.CompleteAuthorization(context.Background(), "auth_code", state)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if token.AccessToken != "access" {
		t.Errorf("expected access, got %s", token.AccessToken)
	}
	if !repo.saveCalled {
		t.Error("expected token to be saved")
	}
}

func TestOAuthUseCase_CompleteAuthorization_StateMismatch(t *testing.T) {
	repo := &mockTokenRepository{}
	provider := &mockOAuthProvider{}
	uc := NewOAuthUseCase(repo, provider)

	uc.StartAuthorization()
	_, err := uc.CompleteAuthorization(context.Background(), "auth_code", "wrong_state")

	if err != ErrStateMismatch {
		t.Errorf("expected ErrStateMismatch, got %v", err)
	}
}

func TestOAuthUseCase_CompleteAuthorization_ExchangeFails(t *testing.T) {
	repo := &mockTokenRepository{}
	provider := &mockOAuthProvider{exchangeErr: errors.New("exchange failed")}
	uc := NewOAuthUseCase(repo, provider)

	_, state := uc.StartAuthorization()
	token, err := uc.CompleteAuthorization(context.Background(), "auth_code", state)

	if err != ErrExchangeFailed {
		t.Errorf("expected ErrExchangeFailed, got %v", err)
	}
	if token != nil {
		t.Error("expected nil token")
	}
}
