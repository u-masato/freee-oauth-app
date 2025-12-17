// freee OAuth認可アプリケーション
//
// クリーンアーキテクチャ/DDDを採用したfreee APIのOAuth認可フロー実装
//
// 使用方法:
//
//	export FREEE_CLIENT_ID="your-client-id"
//	export FREEE_CLIENT_SECRET="your-client-secret"
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"freee-oauth-app/domain"
	"freee-oauth-app/infrastructure/freee"
	"freee-oauth-app/infrastructure/persistence"
	httphandler "freee-oauth-app/interface/http"
	"freee-oauth-app/usecase"
)

const (
	callbackPort = "8080"
	callbackPath = "/callback"
	tokenFile    = "token.json"
)

func main() {
	// 設定の読み込み
	config := loadConfig()

	// 依存性の注入（DI）
	app := initializeApp(config)

	// アプリケーションの実行
	if err := app.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

// Config はアプリケーション設定
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	TokenFile    string
}

func loadConfig() *Config {
	clientID := os.Getenv("FREEE_CLIENT_ID")
	clientSecret := os.Getenv("FREEE_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("FREEE_CLIENT_ID and FREEE_CLIENT_SECRET must be set")
	}

	return &Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  fmt.Sprintf("http://localhost:%s%s", callbackPort, callbackPath),
		TokenFile:    tokenFile,
	}
}

// App はアプリケーションのルートコンポーネント
type App struct {
	oauthUseCase *usecase.OAuthUseCase
	tokenRepo    domain.TokenRepository
}

func initializeApp(config *Config) *App {
	// Infrastructure層の初期化
	tokenRepo := persistence.NewFileTokenRepository(config.TokenFile)
	oauthProvider := freee.NewFreeeOAuthProvider(
		config.ClientID,
		config.ClientSecret,
		config.RedirectURL,
	)

	// UseCase層の初期化
	oauthUseCase := usecase.NewOAuthUseCase(tokenRepo, oauthProvider)

	return &App{
		oauthUseCase: oauthUseCase,
		tokenRepo:    tokenRepo,
	}
}

// Run はアプリケーションを実行する
func (app *App) Run() error {
	ctx := context.Background()

	// 既存のトークンを確認
	token, err := app.oauthUseCase.GetOrRefreshToken(ctx)
	if err == nil {
		fmt.Printf("Loaded existing valid token\n")
		fmt.Printf("  Access Token: %s\n", token.MaskedAccessToken())
		fmt.Printf("  Expires: %s\n", token.Expiry.Format(time.RFC3339))
		fmt.Println("\nToken is ready for API requests.")
		return nil
	}

	if err == usecase.ErrRefreshFailed {
		fmt.Println("Token refresh failed. Starting new OAuth2 flow...")
	} else {
		fmt.Println("No existing token. Starting OAuth2 flow...")
	}

	// 新規OAuth認可フローの開始
	return app.startOAuthFlow(ctx)
}

func (app *App) startOAuthFlow(ctx context.Context) error {
	// 認可フローの開始
	authURL, _ := app.oauthUseCase.StartAuthorization()

	// コールバック用チャネル
	tokenChan := make(chan *domain.Token, 1)
	errChan := make(chan error, 1)

	// HTTPサーバーの起動
	handler := httphandler.NewCallbackHandler(app.oauthUseCase, tokenChan, errChan)
	server := &http.Server{
		Addr:    ":" + callbackPort,
		Handler: handler,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// 認可URLの表示
	fmt.Println("Visit this URL to authorize the application:")
	fmt.Printf("\n%s\n\n", authURL)
	fmt.Println("Waiting for authorization...")

	// コールバック待機
	var token *domain.Token
	select {
	case token = <-tokenChan:
		fmt.Println("\nAuthorization successful!")
	case err := <-errChan:
		shutdownServer(server)
		return fmt.Errorf("authorization failed: %w", err)
	case <-time.After(5 * time.Minute):
		shutdownServer(server)
		return fmt.Errorf("authorization timeout (5 minutes)")
	}

	// サーバーのシャットダウン
	shutdownServer(server)

	// 結果の表示
	fmt.Printf("\nAccess token obtained successfully\n")
	fmt.Printf("  Access Token: %s\n", token.MaskedAccessToken())
	fmt.Printf("  Expires: %s\n", token.Expiry.Format(time.RFC3339))
	if token.HasRefreshToken() {
		fmt.Printf("  Refresh Token: (available)\n")
	}
	fmt.Printf("\nToken saved to %s\n", tokenFile)
	fmt.Println("\nYou can now use this token to make API requests.")

	return nil
}

func shutdownServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
