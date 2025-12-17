package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"freee-oauth-app/domain"
)

var (
	ErrNoToken        = errors.New("no token available")
	ErrRefreshFailed  = errors.New("token refresh failed")
	ErrStateMismatch  = errors.New("state mismatch")
	ErrExchangeFailed = errors.New("token exchange failed")
)

// OAuthUseCase はOAuth認可フローのユースケースを提供する
type OAuthUseCase struct {
	tokenRepo     domain.TokenRepository
	oauthProvider domain.OAuthProvider
	currentState  string
}

// NewOAuthUseCase は新しいOAuthUseCaseを生成する
func NewOAuthUseCase(tokenRepo domain.TokenRepository, oauthProvider domain.OAuthProvider) *OAuthUseCase {
	return &OAuthUseCase{
		tokenRepo:     tokenRepo,
		oauthProvider: oauthProvider,
	}
}

// GetOrRefreshToken は既存のトークンを取得し、必要に応じてリフレッシュする
func (uc *OAuthUseCase) GetOrRefreshToken(ctx context.Context) (*domain.Token, error) {
	token, err := uc.tokenRepo.Load(ctx)
	if err != nil {
		return nil, ErrNoToken
	}

	if token.IsValid() {
		return token, nil
	}

	if token.NeedsRefresh() {
		newToken, err := uc.oauthProvider.Refresh(ctx, token)
		if err != nil {
			return nil, ErrRefreshFailed
		}
		if err := uc.tokenRepo.Save(ctx, newToken); err != nil {
			return nil, err
		}
		return newToken, nil
	}

	return nil, ErrNoToken
}

// StartAuthorization は認可フローを開始し、認可URLとstateを返す
func (uc *OAuthUseCase) StartAuthorization() (authURL string, state string) {
	uc.currentState = generateState()
	authURL = uc.oauthProvider.AuthorizationURL(uc.currentState)
	return authURL, uc.currentState
}

// CompleteAuthorization は認可コードをトークンに交換して保存する
func (uc *OAuthUseCase) CompleteAuthorization(ctx context.Context, code, state string) (*domain.Token, error) {
	if state != uc.currentState {
		return nil, ErrStateMismatch
	}

	token, err := uc.oauthProvider.Exchange(ctx, code)
	if err != nil {
		return nil, ErrExchangeFailed
	}

	if err := uc.tokenRepo.Save(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
