package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"ostenia/internal/config"
	"path/filepath"
	"sync"
	"time"

	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/ssh"
	"path"
	"runtime"
)

type SSHConnection struct {
	SessionID string
	Client    *goph.Client
	SFTP      *sftp.Client
	Shell     io.WriteCloser
	PTY       *ssh.Session
	Context   context.Context
	Cancel    context.CancelFunc
	LastSync  time.Time
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

	// If already connected, do nothing
	if _, ok := m.connections[session.ID]; ok {
		return nil
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
		Callback: ssh.InsecureIgnoreHostKey(), // For simplicity, in production you'd want host key verification
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
	conn.PTY = sshSession

	stdout, _ := sshSession.StdoutPipe()
	stderr, _ := sshSession.StderrPipe()
	stdin, _ := sshSession.StdinPipe()
	conn.Shell = stdin

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
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
				wruntime.EventsEmit(m.ctx, "ssh_output", map[string]interface{}{
					"sessionId": conn.SessionID,
					"data":      data,
				})

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
				wruntime.EventsEmit(m.ctx, "ssh_output", map[string]interface{}{
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
		wruntime.EventsEmit(m.ctx, "ssh_disconnected", conn.SessionID)
		m.mu.Lock()
		delete(m.connections, conn.SessionID)
		m.mu.Unlock()
	}
}

func (m *SSHManager) ResizeTerminal(sessionID string, cols int, rows int) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.PTY == nil {
		return fmt.Errorf("session not found")
	}

	return conn.PTY.WindowChange(rows, cols)
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

func (m *SSHManager) ListFiles(sessionID string, pathStr string) ([]RemoteFile, error) {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.SFTP == nil {
		return nil, fmt.Errorf("session not found or SFTP not connected")
	}

	if pathStr == "" {
		pathStr = "."
	}

	entries, err := conn.SFTP.ReadDir(pathStr)
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

func (m *SSHManager) ExecuteSFTPAction(sessionID string, action string, pathStr string, target string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.SFTP == nil {
		return fmt.Errorf("session not found or SFTP not connected")
	}

	pathStr = path.Clean(pathStr)
	if target != "" {
		target = path.Clean(target)
	}

	switch action {
	case "delete":
		// Check if it's a directory
		info, err := conn.SFTP.Stat(pathStr)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return conn.SFTP.RemoveAll(pathStr)
		}
		return conn.SFTP.Remove(pathStr)
	case "rename":
		return conn.SFTP.Rename(pathStr, target)
	case "mkdir":
		return conn.SFTP.Mkdir(pathStr)
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

	remotePath = path.Clean(remotePath)

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

	// Use path.Join for remote paths
	remotePath = path.Clean(remotePath)

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

	remotePath = path.Clean(remotePath)

	tempDir := filepath.Join(os.TempDir(), "ostenia-ssh-edit")
	os.MkdirAll(tempDir, 0755)

	fileName := path.Base(remotePath)
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

	pathStr, err := conn.SFTP.Getwd()
	return pathStr, err
}
