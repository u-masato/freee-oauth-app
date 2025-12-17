package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"freee-oauth-app/domain"
	"freee-oauth-app/usecase"
)

// モックOAuthUseCase
type mockOAuthUseCase struct {
	getOrRefreshToken func() (*domain.Token, error)
	startAuth         func() (string, string)
	completeAuth      func(ctx context.Context, code, state string) (*domain.Token, error)
}

func (m *mockOAuthUseCase) GetOrRefreshToken(ctx context.Context) (*domain.Token, error) {
	if m.getOrRefreshToken != nil {
		return m.getOrRefreshToken()
	}
	return nil, nil
}

func (m *mockOAuthUseCase) StartAuthorization() (string, string) {
	if m.startAuth != nil {
		return m.startAuth()
	}
	return "", ""
}

func (m *mockOAuthUseCase) CompleteAuthorization(ctx context.Context, code, state string) (*domain.Token, error) {
	if m.completeAuth != nil {
		return m.completeAuth(ctx, code, state)
	}
	return nil, nil
}

func TestCallbackHandler_Success(t *testing.T) {
	token := domain.NewToken("access", "refresh", time.Now().Add(time.Hour))
	tokenChan := make(chan *domain.Token, 1)
	errChan := make(chan error, 1)

	mock := &mockOAuthUseCase{
		completeAuth: func(ctx context.Context, code, state string) (*domain.Token, error) {
			if code != "auth_code" {
				t.Errorf("expected code 'auth_code', got '%s'", code)
			}
			if state != "test_state" {
				t.Errorf("expected state 'test_state', got '%s'", state)
			}
			return token, nil
		},
	}

	handler := NewCallbackHandler(mock, tokenChan, errChan)

	req := httptest.NewRequest("GET", "/callback?code=auth_code&state=test_state", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	select {
	case received := <-tokenChan:
		if received.AccessToken != "access" {
			t.Errorf("expected access token 'access', got '%s'", received.AccessToken)
		}
	default:
		t.Error("expected token to be sent to channel")
	}
}

func TestCallbackHandler_ErrorParam(t *testing.T) {
	tokenChan := make(chan *domain.Token, 1)
	errChan := make(chan error, 1)
	mock := &mockOAuthUseCase{}

	handler := NewCallbackHandler(mock, tokenChan, errChan)

	req := httptest.NewRequest("GET", "/callback?error=access_denied&error_description=User+denied", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	select {
	case err := <-errChan:
		if err == nil {
			t.Error("expected error to be sent to channel")
		}
	default:
		t.Error("expected error to be sent to channel")
	}
}

func TestCallbackHandler_StateMismatch(t *testing.T) {
	tokenChan := make(chan *domain.Token, 1)
	errChan := make(chan error, 1)

	mock := &mockOAuthUseCase{
		completeAuth: func(ctx context.Context, code, state string) (*domain.Token, error) {
			return nil, usecase.ErrStateMismatch
		},
	}

	handler := NewCallbackHandler(mock, tokenChan, errChan)

	req := httptest.NewRequest("GET", "/callback?code=auth_code&state=wrong_state", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCallbackHandler_ExchangeFails(t *testing.T) {
	tokenChan := make(chan *domain.Token, 1)
	errChan := make(chan error, 1)

	mock := &mockOAuthUseCase{
		completeAuth: func(ctx context.Context, code, state string) (*domain.Token, error) {
			return nil, errors.New("exchange failed")
		},
	}

	handler := NewCallbackHandler(mock, tokenChan, errChan)

	req := httptest.NewRequest("GET", "/callback?code=auth_code&state=test_state", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestCallbackHandler_NoCode(t *testing.T) {
	tokenChan := make(chan *domain.Token, 1)
	errChan := make(chan error, 1)
	mock := &mockOAuthUseCase{}

	handler := NewCallbackHandler(mock, tokenChan, errChan)

	req := httptest.NewRequest("GET", "/callback?state=test_state", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
