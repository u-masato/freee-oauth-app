package persistence

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"freee-oauth-app/domain"
)

func TestFileTokenRepository_Save_And_Load(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "token.json")
	repo := NewFileTokenRepository(filePath)

	ctx := context.Background()
	expiry := time.Now().Add(time.Hour).Truncate(time.Second)
	token := domain.NewToken("access123", "refresh456", expiry)

	// Save
	err := repo.Save(ctx, token)
	if err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("token file was not created")
	}

	// Load
	loaded, err := repo.Load(ctx)
	if err != nil {
		t.Fatalf("failed to load token: %v", err)
	}

	if loaded.AccessToken != "access123" {
		t.Errorf("expected access123, got %s", loaded.AccessToken)
	}
	if loaded.RefreshToken != "refresh456" {
		t.Errorf("expected refresh456, got %s", loaded.RefreshToken)
	}
	if !loaded.Expiry.Equal(expiry) {
		t.Errorf("expected expiry %v, got %v", expiry, loaded.Expiry)
	}
}

func TestFileTokenRepository_Load_WhenFileNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "nonexistent.json")
	repo := NewFileTokenRepository(filePath)

	ctx := context.Background()
	_, err := repo.Load(ctx)

	if err == nil {
		t.Error("expected error when file does not exist")
	}
}

func TestFileTokenRepository_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "token.json")
	repo := NewFileTokenRepository(filePath)

	ctx := context.Background()

	// Before save
	if repo.Exists(ctx) {
		t.Error("expected Exists to return false before save")
	}

	// After save
	token := domain.NewToken("access", "refresh", time.Now().Add(time.Hour))
	repo.Save(ctx, token)

	if !repo.Exists(ctx) {
		t.Error("expected Exists to return true after save")
	}
}

func TestFileTokenRepository_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "token.json")
	repo := NewFileTokenRepository(filePath)

	ctx := context.Background()
	token := domain.NewToken("access", "refresh", time.Now().Add(time.Hour))
	repo.Save(ctx, token)

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	// ファイルのパーミッションが0600であることを確認
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected file permission 0600, got %o", perm)
	}
}
