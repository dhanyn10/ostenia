package usecase_test

import (
	"context"
	"testing"

	"ostenia/internal/domain"
	"ostenia/internal/infrastructure"
	"ostenia/internal/usecase"
)

func TestConfigUsecase(t *testing.T) {
	adapter := infrastructure.NewConfigAdapter()
	uc := usecase.NewConfigUsecase(adapter)

	baseDir := uc.GetBaseDir()
	if baseDir == "" {
		t.Errorf("expected non-empty base directory")
	}

	cfg, err := uc.GetConfig()
	if err != nil {
		t.Fatalf("unexpected error getting config: %v", err)
	}

	if cfg == nil {
		t.Fatalf("expected non-nil config")
	}

	cfg.PHPVersion = "8.2"
	err = uc.SaveConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error saving config: %v", err)
	}

	updated, err := uc.GetConfig()
	if err != nil || updated.PHPVersion != "8.2" {
		t.Errorf("expected updated PHP version to be 8.2, got %s", updated.PHPVersion)
	}
}

type mockSSHManager struct{}

func (m *mockSSHManager) GetSessions() ([]domain.SSHSession, error) {
	return []domain.SSHSession{}, nil
}
func (m *mockSSHManager) SaveSessions(sessions []domain.SSHSession) error { return nil }
func (m *mockSSHManager) Connect(ctx context.Context, session domain.SSHSession) error {
	return nil
}
func (m *mockSSHManager) GetCurrentPath(sessionID string) (string, error) {
	return "/var/www", nil
}
func (m *mockSSHManager) ExecuteSFTPAction(sessionID, action, path, newPath string) error {
	return nil
}

func TestSSHUsecase(t *testing.T) {
	mock := &mockSSHManager{}
	uc := usecase.NewSSHUsecase(mock)

	sess, err := uc.GetSessions()
	if err != nil {
		t.Fatalf("unexpected error getting sessions: %v", err)
	}
	if len(sess) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sess))
	}

	path, err := uc.GetCurrentPath("test-session")
	if err != nil || path != "/var/www" {
		t.Errorf("expected /var/www, got %s", path)
	}

	err = uc.ExecuteSFTPAction("test-session", "mkdir", "/var/www/test", "")
	if err != nil {
		t.Errorf("unexpected error in sftp action: %v", err)
	}
}
