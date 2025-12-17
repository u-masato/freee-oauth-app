package freee

import (
	"context"
	"time"

	"freee-oauth-app/domain"

	"golang.org/x/oauth2"
)

const (
	defaultAuthURL  = "https://accounts.secure.freee.co.jp/public_api/authorize"
	defaultTokenURL = "https://accounts.secure.freee.co.jp/public_api/token"
)

// FreeeOAuthProvider はfreee APIのOAuth認可プロバイダー
type FreeeOAuthProvider struct {
	config *oauth2.Config
}

// NewFreeeOAuthProvider は新しいFreeeOAuthProviderを生成する
func NewFreeeOAuthProvider(clientID, clientSecret, redirectURL string) *FreeeOAuthProvider {
	return NewFreeeOAuthProviderWithEndpoint(clientID, clientSecret, redirectURL, defaultAuthURL, defaultTokenURL)
}

// NewFreeeOAuthProviderWithEndpoint はカスタムエンドポイントでFreeeOAuthProviderを生成する
func NewFreeeOAuthProviderWithEndpoint(clientID, clientSecret, redirectURL, authURL, tokenURL string) *FreeeOAuthProvider {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"read", "write"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}

	return &FreeeOAuthProvider{
		config: config,
	}
}

// AuthorizationURL は認可URLを生成する
func (p *FreeeOAuthProvider) AuthorizationURL(state string) string {
	return p.config.AuthCodeURL(state)
}

// Exchange は認可コードをトークンに交換する
func (p *FreeeOAuthProvider) Exchange(ctx context.Context, code string) (*domain.Token, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return domain.NewToken(token.AccessToken, token.RefreshToken, token.Expiry), nil
}

// Refresh はリフレッシュトークンを使用してトークンを更新する
func (p *FreeeOAuthProvider) Refresh(ctx context.Context, token *domain.Token) (*domain.Token, error) {
	oauth2Token := &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       time.Now().Add(-time.Hour), // 期限切れとしてマーク
	}

	tokenSource := p.config.TokenSource(ctx, oauth2Token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}

	return domain.NewToken(newToken.AccessToken, newToken.RefreshToken, newToken.Expiry), nil
}
