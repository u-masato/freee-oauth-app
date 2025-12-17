package persistence

import (
	"context"
	"os"

	"freee-oauth-app/domain"

	"github.com/u-masato/freee-api-go/auth"
)

// FileTokenRepository はファイルベースのトークンリポジトリ
type FileTokenRepository struct {
	filePath string
}

// NewFileTokenRepository は新しいFileTokenRepositoryを生成する
func NewFileTokenRepository(filePath string) *FileTokenRepository {
	return &FileTokenRepository{
		filePath: filePath,
	}
}

// Save はトークンをファイルに保存する
func (r *FileTokenRepository) Save(ctx context.Context, token *domain.Token) error {
	return auth.SaveTokenToFile(token.ToOAuth2Token(), r.filePath)
}

// Load はファイルからトークンを読み込む
func (r *FileTokenRepository) Load(ctx context.Context) (*domain.Token, error) {
	oauth2Token, err := auth.LoadTokenFromFile(r.filePath)
	if err != nil {
		return nil, err
	}

	return domain.FromOAuth2Token(oauth2Token), nil
}

// Exists はトークンファイルが存在するかを確認する
func (r *FileTokenRepository) Exists(ctx context.Context) bool {
	_, err := os.Stat(r.filePath)
	return err == nil
}
