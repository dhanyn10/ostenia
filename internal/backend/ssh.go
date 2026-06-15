package backend

import (
	"ostenia/internal/config"
	"ostenia/internal/service"
)

type SSHManagerDelegate struct {
	SSHManager *service.SSHManager
	Config     *config.Config
}

func (s *SSHManagerDelegate) Connect(session config.SSHSession) error {
	return s.SSHManager.Connect(session)
}

func (s *SSHManagerDelegate) Disconnect(sessionID string) {
	s.SSHManager.Disconnect(sessionID)
}

func (s *SSHManagerDelegate) SendInput(sessionID string, data string) error {
	return s.SSHManager.SendInput(sessionID, data)
}

func (s *SSHManagerDelegate) ResizeTerminal(sessionID string, cols int, rows int) error {
	return s.SSHManager.ResizeTerminal(sessionID, cols, rows)
}

func (s *SSHManagerDelegate) ListFiles(sessionID string, path string) ([]service.RemoteFile, error) {
	return s.SSHManager.ListFiles(sessionID, path)
}

func (s *SSHManagerDelegate) ExecuteSFTPAction(sessionID string, action string, path string, target string) error {
	return s.SSHManager.ExecuteSFTPAction(sessionID, action, path, target)
}

func (s *SSHManagerDelegate) EditFile(sessionID string, remotePath string) error {
	return s.SSHManager.EditFile(sessionID, remotePath, s.Config.DefaultEditor)
}

func (s *SSHManagerDelegate) GetCurrentPath(sessionID string) (string, error) {
	return s.SSHManager.GetCurrentPath(sessionID)
}

func (s *SSHManagerDelegate) DownloadFile(sessionID string, remotePath string, localPath string) error {
	return s.SSHManager.DownloadFile(sessionID, remotePath, localPath)
}

func (s *SSHManagerDelegate) UploadFile(sessionID string, localPath string, remotePath string) error {
	return s.SSHManager.UploadFile(sessionID, localPath, remotePath)
}
