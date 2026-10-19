package usecase

import (
	"context"

	"ostenia/internal/domain"
)

// ConfigUsecase handles domain business logic for application configuration.
type ConfigUsecase struct {
	repo domain.ConfigRepository
}

func NewConfigUsecase(repo domain.ConfigRepository) *ConfigUsecase {
	return &ConfigUsecase{repo: repo}
}

func (u *ConfigUsecase) GetConfig() (*domain.Config, error) {
	return u.repo.LoadConfig()
}

func (u *ConfigUsecase) SaveConfig(cfg *domain.Config) error {
	return u.repo.SaveConfig(cfg)
}

func (u *ConfigUsecase) GetBaseDir() string {
	return u.repo.GetBaseDir()
}

// SSHUsecase orchestrates SSH profiles, SFTP file management, and terminal session handling.
type SSHUsecase struct {
	sshManager domain.SSHManager
}

func NewSSHUsecase(mgr domain.SSHManager) *SSHUsecase {
	return &SSHUsecase{
		sshManager: mgr,
	}
}

func (u *SSHUsecase) GetSessions() ([]domain.SSHSession, error) {
	return u.sshManager.GetSessions()
}

func (u *SSHUsecase) SaveSessions(sessions []domain.SSHSession) error {
	return u.sshManager.SaveSessions(sessions)
}

func (u *SSHUsecase) Connect(ctx context.Context, session domain.SSHSession) error {
	return u.sshManager.Connect(ctx, session)
}

func (u *SSHUsecase) GetCurrentPath(sessionID string) (string, error) {
	return u.sshManager.GetCurrentPath(sessionID)
}

func (u *SSHUsecase) ExecuteSFTPAction(sessionID, action, path, newPath string) error {
	return u.sshManager.ExecuteSFTPAction(sessionID, action, path, newPath)
}

// PluginUsecase orchestrates plugin discovery and management.
type PluginUsecase struct{}

func NewPluginUsecase() *PluginUsecase {
	return &PluginUsecase{}
}

// ServiceUsecase manages server stack services (PHP, Apache, Nginx, MySQL, Node.js, Python, OpenSSL, HeidiSQL).
type ServiceUsecase struct {
	orchestrator domain.Orchestrator
}

func NewServiceUsecase(orch domain.Orchestrator) *ServiceUsecase {
	return &ServiceUsecase{
		orchestrator: orch,
	}
}

func (u *ServiceUsecase) IsRunning(name string) bool {
	return u.orchestrator.IsRunning(name)
}

func (u *ServiceUsecase) GetDetailedInfo(name string) domain.ServiceDetailedInfo {
	return u.orchestrator.GetDetailedInfo(name)
}

func (u *ServiceUsecase) StartService(ctx context.Context, name, binaryPath string, args []string, workingDir string) error {
	return u.orchestrator.StartService(ctx, name, binaryPath, args, workingDir)
}

func (u *ServiceUsecase) StopService(ctx context.Context, name string) error {
	return u.orchestrator.StopService(ctx, name)
}
