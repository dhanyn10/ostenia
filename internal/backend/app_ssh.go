package backend

import (
	"ostenia/internal/config"
	"ostenia/internal/backend/interfaces"
	"path/filepath"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// GetSSHSessions returns the list of saved SSH sessions
func (a *App) GetSSHSessions() ([]config.SSHSession, error) {
	return config.LoadSSHSessions()
}

// SaveSSHSessions saves the entire list of SSH sessions
func (a *App) SaveSSHSessions(sessions []config.SSHSession) error {
	return config.SaveSSHSessions(sessions)
}

// AddSSHSession adds a new SSH session
func (a *App) AddSSHSession(session config.SSHSession) error {
	return config.AddSSHSession(session)
}

// UpdateSSHSession updates an existing SSH session
func (a *App) UpdateSSHSession(session config.SSHSession) error {
	return config.UpdateSSHSession(session)
}

// DeleteSSHSession removes an SSH session by ID
func (a *App) DeleteSSHSession(id string) error {
	return config.DeleteSSHSession(id)
}

// ConnectSSH initiates an SSH connection
func (a *App) ConnectSSH(session config.SSHSession) error {
	return a.sshManager.Connect(session)
}

// DisconnectSSH closes an SSH connection
func (a *App) DisconnectSSH(sessionID string) {
	a.sshManager.Disconnect(sessionID)
}

// SendSSHInput sends terminal input to an active SSH session
func (a *App) SendSSHInput(sessionID string, data string) error {
	return a.sshManager.SendInput(sessionID, data)
}

// ResizeSSHTerminal updates the PTY size for an SSH session
func (a *App) ResizeSSHTerminal(sessionID string, cols int, rows int) error {
	return a.sshManager.ResizeTerminal(sessionID, cols, rows)
}

// GetRemoteFiles lists files in a remote directory via SFTP
func (a *App) GetRemoteFiles(sessionID string, path string) ([]interfaces.RemoteFile, error) {
	return a.sshManager.ListFiles(sessionID, path)
}

// ExecuteSFTPAction performs file operations (rename, delete, mkdir) on a remote host
func (a *App) ExecuteSFTPAction(sessionID string, action string, path string, target string) error {
	return a.sshManager.ExecuteSFTPAction(sessionID, action, path, target)
}

// EditRemoteFile downloads a remote file to a temporary location and opens it in the default editor
func (a *App) EditRemoteFile(sessionID string, remotePath string) error {
	return a.sshManager.EditFile(sessionID, remotePath, a.cfg.DefaultEditor)
}

// GetRemoteCurrentPath returns the current working directory of an SSH session
func (a *App) GetRemoteCurrentPath(sessionID string) (string, error) {
	return a.sshManager.GetCurrentPath(sessionID)
}

// DownloadRemoteFile downloads a file from a remote host to the local machine
func (a *App) DownloadRemoteFile(sessionID string, remotePath string) error {
	fileName := filepath.Base(remotePath)
	localPath, err := a.runtime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title:           "Download File",
		DefaultFilename: fileName,
	})
	if err != nil || localPath == "" {
		return err
	}
	return a.sshManager.DownloadFile(sessionID, remotePath, localPath)
}

// UploadRemoteFile uploads a file from the local machine to a remote host
func (a *App) UploadRemoteFile(sessionID string, remoteDir string) error {
	localPath, err := a.runtime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Upload File",
	})
	if err != nil || localPath == "" {
		return err
	}
	remotePath := filepath.ToSlash(filepath.Join(remoteDir, filepath.Base(localPath)))
	return a.sshManager.UploadFile(sessionID, localPath, remotePath)
}
