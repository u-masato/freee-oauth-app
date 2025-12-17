package domain

import "context"

// TokenRepository はトークンの永続化を担当するリポジトリのインターフェース
type TokenRepository interface {
	Save(ctx context.Context, token *Token) error
	Load(ctx context.Context) (*Token, error)
	Exists(ctx context.Context) bool
}

// OAuthProvider はOAuth認可フローを担当するプロバイダーのインターフェース
type OAuthProvider interface {
	// AuthorizationURL は認可URLを生成する
	AuthorizationURL(state string) string
	// Exchange は認可コードをトークンに交換する
	Exchange(ctx context.Context, code string) (*Token, error)
	// Refresh はリフレッシュトークンを使用してトークンを更新する
	Refresh(ctx context.Context, token *Token) (*Token, error)
}
