package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"ostenia/internal/config"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type SSHConnection struct {
	SessionID string
	Client    *goph.Client
	SFTP      *sftp.Client
	Shell     io.WriteCloser
	Context   context.Context
	Cancel    context.CancelFunc
}

type SSHManager struct {
	ctx         context.Context
	connections map[string]*SSHConnection
	mu          sync.RWMutex
}

func NewSSHManager(ctx context.Context) *SSHManager {
	return &SSHManager{
		ctx:         ctx,
		connections: make(map[string]*SSHConnection),
	}
}

func (m *SSHManager) Connect(session config.SSHSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close existing if any
	if conn, ok := m.connections[session.ID]; ok {
		conn.Cancel()
		if conn.SFTP != nil {
			conn.SFTP.Close()
		}
		conn.Client.Close()
		delete(m.connections, session.ID)
	}

	var auth goph.Auth
	var err error

	if session.AuthMethod == "password" {
		auth = goph.Password(session.Password)
	} else {
		if session.Passphrase != "" {
			auth, err = goph.Key(session.KeyPath, session.Passphrase)
		} else {
			auth, err = goph.Key(session.KeyPath, "")
		}
		if err != nil {
			return err
		}
	}

	client, err := goph.NewConn(&goph.Config{
		User:     session.User,
		Addr:     session.Host,
		Port:     uint(session.Port),
		Auth:     auth,
		Timeout:  10 * time.Second,
		Callback: goph.InsecureIgnoreHostKey(), // For simplicity, in production you'd want host key verification
	})

	if err != nil {
		return err
	}

	sftpClient, err := client.NewSftp()
	if err != nil {
		client.Close()
		return err
	}

	ctx, cancel := context.WithCancel(m.ctx)
	conn := &SSHConnection{
		SessionID: session.ID,
		Client:    client,
		SFTP:      sftpClient,
		Context:   ctx,
		Cancel:    cancel,
	}

	m.connections[session.ID] = conn

	// Start terminal session
	go m.startTerminal(conn)

	return nil
}

func (m *SSHManager) startTerminal(conn *SSHConnection) {
	sshSession, err := conn.Client.NewSession()
	if err != nil {
		fmt.Printf("Failed to create SSH session: %v\n", err)
		return
	}
	defer sshSession.Close()

	stdout, _ := sshSession.StdoutPipe()
	stderr, _ := sshSession.StderrPipe()
	stdin, _ := sshSession.StdinPipe()
	conn.Shell = stdin

	modes := goph.TerminalModes{
		goph.ECHO:          1,
		goph.TTY_OP_ISPEED: 14400,
		goph.TTY_OP_OSPEED: 14400,
	}

	if err := sshSession.RequestPty("xterm-256color", 80, 40, modes); err != nil {
		fmt.Printf("request for pseudo terminal failed: %v\n", err)
		return
	}

	if err := sshSession.Shell(); err != nil {
		fmt.Printf("failed to start shell: %v\n", err)
		return
	}

	// Channel to signal shell exit
	exitChan := make(chan struct{})

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				data := string(buf[:n])
				runtime.EventsEmit(m.ctx, "ssh_output", map[string]interface{}{
					"sessionId": conn.SessionID,
					"data":      data,
				})

				// Heuristic to detect directory change:
				// If we see a common prompt pattern, we try to get the real PWD from the shell.
				if strings.Contains(data, "$") || strings.Contains(data, "#") || strings.Contains(data, ">") {
					go func() {
						time.Sleep(150 * time.Millisecond) // Wait for command to finish and prompt to appear

						// Create a one-off session to get the current working directory of the user's shell.
						// Note: This assumes a Unix-like environment for SSH.
						cmdSess, err := conn.Client.NewSession()
						if err == nil {
							defer cmdSess.Close()
							out, err := cmdSess.Output("pwd")
							if err == nil {
								path := strings.TrimSpace(string(out))
								if path != "" {
									runtime.EventsEmit(m.ctx, "ssh_path_changed", map[string]interface{}{
										"sessionId": conn.SessionID,
										"path":      path,
									})
								}
							}
						}
					}()
				}
			}
			if err != nil {
				break
			}
		}
		close(exitChan)
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				runtime.EventsEmit(m.ctx, "ssh_output", map[string]interface{}{
					"sessionId": conn.SessionID,
					"data":      string(buf[:n]),
				})
			}
			if err != nil {
				break
			}
		}
	}()

	select {
	case <-conn.Context.Done():
		sshSession.Close()
	case <-exitChan:
		runtime.EventsEmit(m.ctx, "ssh_disconnected", conn.SessionID)
		m.mu.Lock()
		delete(m.connections, conn.SessionID)
		m.mu.Unlock()
	}
}

func (m *SSHManager) SendInput(sessionID string, data string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.Shell == nil {
		return fmt.Errorf("session not found or not connected")
	}

	_, err := conn.Shell.Write([]byte(data))
	return err
}

func (m *SSHManager) Disconnect(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, ok := m.connections[sessionID]; ok {
		conn.Cancel()
		if conn.SFTP != nil {
			conn.SFTP.Close()
		}
		conn.Client.Close()
		delete(m.connections, sessionID)
	}
}

type RemoteFile struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"isDir"`
	ModTime int64  `json:"modTime"`
	Mode    string `json:"mode"`
}

func (m *SSHManager) ListFiles(sessionID string, path string) ([]RemoteFile, error) {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.SFTP == nil {
		return nil, fmt.Errorf("session not found or SFTP not connected")
	}

	if path == "" {
		path = "."
	}

	entries, err := conn.SFTP.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var files []RemoteFile
	for _, entry := range entries {
		files = append(files, RemoteFile{
			Name:    entry.Name(),
			Size:    entry.Size(),
			IsDir:   entry.IsDir(),
			ModTime: entry.ModTime().Unix(),
			Mode:    entry.Mode().String(),
		})
	}

	return files, nil
}

func (m *SSHManager) ExecuteSFTPAction(sessionID string, action string, path string, target string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.SFTP == nil {
		return fmt.Errorf("session not found or SFTP not connected")
	}

	switch action {
	case "delete":
		// Check if it's a directory
		info, err := conn.SFTP.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return conn.SFTP.RemoveAll(path)
		}
		return conn.SFTP.Remove(path)
	case "rename":
		return conn.SFTP.Rename(path, target)
	case "mkdir":
		return conn.SFTP.Mkdir(path)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func (m *SSHManager) DownloadFile(sessionID string, remotePath string, localPath string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.SFTP == nil {
		return fmt.Errorf("session not found or SFTP not connected")
	}

	src, err := conn.SFTP.Open(remotePath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func (m *SSHManager) UploadFile(sessionID string, localPath string, remotePath string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.SFTP == nil {
		return fmt.Errorf("session not found or SFTP not connected")
	}

	src, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := conn.SFTP.Create(remotePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func (m *SSHManager) EditFile(sessionID string, remotePath string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.SFTP == nil {
		return fmt.Errorf("session not found or SFTP not connected")
	}

	tempDir := filepath.Join(os.TempDir(), "ostenia-ssh-edit")
	os.MkdirAll(tempDir, 0755)

	fileName := filepath.Base(remotePath)
	localPath := filepath.Join(tempDir, fmt.Sprintf("%d-%s", time.Now().Unix(), fileName))

	err := m.DownloadFile(sessionID, remotePath, localPath)
	if err != nil {
		return err
	}

	// Get initial file info
	initialInfo, _ := os.Stat(localPath)

	// Open with default editor
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("notepad.exe", localPath)
	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command("open", "-t", localPath)
	} else {
		// Try to find a common editor on Linux
		editors := []string{"xdg-open", "gnome-text-editor", "kwrite", "gedit", "nano"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				cmd = exec.Command(e, localPath)
				break
			}
		}
	}

	if cmd == nil {
		return fmt.Errorf("no suitable text editor found")
	}

	err = cmd.Run()
	if err != nil {
		return err
	}

	// Check if file was modified
	finalInfo, err := os.Stat(localPath)
	if err == nil && finalInfo.ModTime().After(initialInfo.ModTime()) {
		return m.UploadFile(sessionID, localPath, remotePath)
	}

	return nil
}

func (m *SSHManager) GetCurrentPath(sessionID string) (string, error) {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.SFTP == nil {
		return "", fmt.Errorf("session not found or SFTP not connected")
	}

	path, err := conn.SFTP.Getwd()
	return path, err
}
