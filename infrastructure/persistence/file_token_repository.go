package persistence

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"freee-oauth-app/domain"
)

// tokenJSON はトークンのJSON表現
type tokenJSON struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

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
	data := tokenJSON{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.filePath, jsonData, 0600)
}

// Load はファイルからトークンを読み込む
func (r *FileTokenRepository) Load(ctx context.Context) (*domain.Token, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, err
	}

	var tokenData tokenJSON
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return nil, err
	}

	return domain.NewToken(tokenData.AccessToken, tokenData.RefreshToken, tokenData.Expiry), nil
}

// Exists はトークンファイルが存在するかを確認する
func (r *FileTokenRepository) Exists(ctx context.Context) bool {
	_, err := os.Stat(r.filePath)
	return err == nil
}
