package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"freee-oauth-app/domain"
	"freee-oauth-app/usecase"
)

// OAuthUseCaseInterface はHTTPハンドラが必要とするユースケースのインターフェース
type OAuthUseCaseInterface interface {
	GetOrRefreshToken(ctx context.Context) (*domain.Token, error)
	StartAuthorization() (authURL string, state string)
	CompleteAuthorization(ctx context.Context, code, state string) (*domain.Token, error)
}

// CallbackHandler はOAuthコールバックを処理するHTTPハンドラ
type CallbackHandler struct {
	useCase   OAuthUseCaseInterface
	tokenChan chan<- *domain.Token
	errChan   chan<- error
}

// NewCallbackHandler は新しいCallbackHandlerを生成する
func NewCallbackHandler(useCase OAuthUseCaseInterface, tokenChan chan<- *domain.Token, errChan chan<- error) *CallbackHandler {
	return &CallbackHandler{
		useCase:   useCase,
		tokenChan: tokenChan,
		errChan:   errChan,
	}
}

// ServeHTTP はHTTPリクエストを処理する
func (h *CallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// エラーパラメータのチェック
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		errDesc := r.URL.Query().Get("error_description")
		h.errChan <- fmt.Errorf("%s: %s", errParam, errDesc)
		http.Error(w, "Authorization failed. You can close this window.", http.StatusBadRequest)
		return
	}

	// 認可コードの取得
	code := r.URL.Query().Get("code")
	if code == "" {
		h.errChan <- errors.New("no authorization code received")
		http.Error(w, "No authorization code received.", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")

	// 認可コードをトークンに交換
	token, err := h.useCase.CompleteAuthorization(r.Context(), code, state)
	if err != nil {
		h.errChan <- err
		if errors.Is(err, usecase.ErrStateMismatch) {
			http.Error(w, "Invalid state parameter.", http.StatusBadRequest)
			return
		}
		http.Error(w, "Token exchange failed.", http.StatusInternalServerError)
		return
	}

	h.tokenChan <- token
	h.renderSuccess(w)
}

func (h *CallbackHandler) renderSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
    <title>Authorization Successful</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background-color: #f5f5f5;
        }
        .container {
            text-align: center;
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 { color: #2ecc71; }
        p { color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Authorization Successful</h1>
        <p>You can close this window and return to the terminal.</p>
    </div>
</body>
</html>`)
}
