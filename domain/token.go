package domain

import "time"

const (
	// トークンが無効とみなされる残り時間のしきい値
	tokenExpiryBuffer = 5 * time.Minute
	// マスク表示時の文字数
	maskedTokenLength = 20
)

// Token はOAuthアクセストークンを表す値オブジェクト
type Token struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

// NewToken は新しいTokenを生成する
func NewToken(accessToken, refreshToken string, expiry time.Time) *Token {
	return &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Expiry:       expiry,
	}
}

// IsValid はトークンが有効かどうかを判定する
// 有効期限の5分前から無効とみなす
func (t *Token) IsValid() bool {
	return time.Now().Add(tokenExpiryBuffer).Before(t.Expiry)
}

// NeedsRefresh はトークンのリフレッシュが必要かつ可能かを判定する
func (t *Token) NeedsRefresh() bool {
	return !t.IsValid() && t.HasRefreshToken()
}

// HasRefreshToken はリフレッシュトークンを持っているかを判定する
func (t *Token) HasRefreshToken() bool {
	return t.RefreshToken != ""
}

// MaskedAccessToken はマスクされたアクセストークンを返す
func (t *Token) MaskedAccessToken() string {
	if len(t.AccessToken) <= maskedTokenLength {
		return t.AccessToken
	}
	return t.AccessToken[:maskedTokenLength] + "..."
}
