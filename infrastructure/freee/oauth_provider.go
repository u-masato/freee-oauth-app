package freee

import (
	"context"
	"time"

	"freee-oauth-app/domain"

	"github.com/u-masato/freee-api-go/auth"
	"golang.org/x/oauth2"
)

// FreeeOAuthProvider はfreee APIのOAuth認可プロバイダー
type FreeeOAuthProvider struct {
	config *auth.Config
}

// NewFreeeOAuthProvider は新しいFreeeOAuthProviderを生成する
func NewFreeeOAuthProvider(clientID, clientSecret, redirectURL string) *FreeeOAuthProvider {
	return &FreeeOAuthProvider{
		config: auth.NewConfig(clientID, clientSecret, redirectURL, []string{"read", "write"}),
	}
}

// NewFreeeOAuthProviderWithEndpoint はカスタムエンドポイントでFreeeOAuthProviderを生成する
func NewFreeeOAuthProviderWithEndpoint(clientID, clientSecret, redirectURL, authURL, tokenURL string) *FreeeOAuthProvider {
	return &FreeeOAuthProvider{
		config: auth.NewConfigWithEndpoint(clientID, clientSecret, redirectURL, []string{"read", "write"}, authURL, tokenURL),
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

	return domain.FromOAuth2Token(token), nil
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

	return domain.FromOAuth2Token(newToken), nil
}
