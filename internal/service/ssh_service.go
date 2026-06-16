package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"ostenia/internal/config"
	"ostenia/internal/plugins/utils"
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
	// Set UTF-8 environment for the session
	os.Setenv("LANG", "en_US.UTF-8")
	os.Setenv("LC_ALL", "en_US.UTF-8")

	m.mu.Lock()
	defer m.mu.Unlock()

	// If already connected, do nothing
	if _, ok := m.connections[session.ID]; ok {
		return nil
	}

	auth, err := m.getAuth(session)
	if err != nil {
		return err
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

func (m *SSHManager) getAuth(session config.SSHSession) (goph.Auth, error) {
	if session.AuthMethod == "password" {
		return goph.Password(session.Password), nil
	}
	return goph.Key(session.KeyPath, session.Passphrase)
}

func (m *SSHManager) startTerminal(conn *SSHConnection) {
	sshSession, err := conn.Client.NewSession()
	if err != nil {
		fmt.Printf("Failed to create SSH session: %v\n", err)
		return
	}
	conn.PTY = sshSession

	stdout, _ := sshSession.StdoutPipe()
	stdin, _ := sshSession.StdinPipe()
	conn.Shell = stdin

	if err := m.requestPtyAndShell(sshSession); err != nil {
		return
	}

	exitChan := make(chan struct{})
	go m.pipeOutput(conn.SessionID, stdout, exitChan)

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

func (m *SSHManager) requestPtyAndShell(sshSession *ssh.Session) error {
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
		ssh.ICANON:        0,
	}

	if err := sshSession.RequestPty("xterm-256color", 50, 200, modes); err != nil {
		fmt.Printf("request for pseudo terminal failed: %v\n", err)
		return err
	}

	if err := sshSession.Shell(); err != nil {
		fmt.Printf("failed to start shell: %v\n", err)
		return err
	}
	return nil
}

func (m *SSHManager) pipeOutput(sessionID string, stdout io.Reader, exitChan chan struct{}) {
	buf := make([]byte, 2048)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			wruntime.EventsEmit(m.ctx, "ssh_output", map[string]interface{}{
				"sessionId": sessionID,
				"data":      string(buf[:n]),
			})
		}
		if err != nil {
			select {
			case <-exitChan:
			default:
				close(exitChan)
			}
			break
		}
	}
}

func (m *SSHManager) ResizeTerminal(sessionID string, cols int, rows int) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.PTY == nil {
		return fmt.Errorf("session not found")
	}

	if cols < 100 {
		cols = 100
	}
	if rows < 10 {
		rows = 10
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

func (m *SSHManager) ExecuteSFTPAction(sessionID string, action string, remotePath string, target string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.SFTP == nil {
		return fmt.Errorf("session not found or SFTP not connected")
	}

	remotePath = path.Clean(remotePath)
	if target != "" {
		target = path.Clean(target)
	}

	switch action {
	case "delete":
		return m.sftpDelete(conn, remotePath)
	case "rename":
		return conn.SFTP.Rename(remotePath, target)
	case "mkdir":
		return conn.SFTP.Mkdir(remotePath)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func (m *SSHManager) sftpDelete(conn *SSHConnection, remotePath string) error {
	info, err := conn.SFTP.Stat(remotePath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return conn.SFTP.RemoveAll(remotePath)
	}
	return conn.SFTP.Remove(remotePath)
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

func (m *SSHManager) EditFile(sessionID string, remotePath string, defaultEditor string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok || conn.SFTP == nil {
		return fmt.Errorf("session not found or SFTP not connected")
	}

	remotePath = path.Clean(remotePath)
	localPath, err := m.prepareLocalFileForEdit(sessionID, remotePath)
	if err != nil {
		return err
	}

	initialInfo, _ := os.Stat(localPath)

	if err := m.openEditor(localPath, defaultEditor); err != nil {
		return err
	}

	finalInfo, err := os.Stat(localPath)
	if err == nil && finalInfo.ModTime().After(initialInfo.ModTime()) {
		return m.UploadFile(sessionID, localPath, remotePath)
	}

	return nil
}

func (m *SSHManager) prepareLocalFileForEdit(sessionID, remotePath string) (string, error) {
	tempDir := filepath.Join(os.TempDir(), "ostenia-ssh-edit")
	os.MkdirAll(tempDir, 0755)

	fileName := path.Base(remotePath)
	localPath := filepath.Join(tempDir, fmt.Sprintf("%d-%s", time.Now().Unix(), fileName))

	if err := m.DownloadFile(sessionID, remotePath, localPath); err != nil {
		return "", err
	}
	return localPath, nil
}

func (m *SSHManager) openEditor(localPath, defaultEditor string) error {
	cmd := m.getEditorCommand(localPath, defaultEditor)
	if cmd == nil {
		return fmt.Errorf("no suitable text editor found")
	}
	return cmd.Run()
}

func (m *SSHManager) getEditorCommand(localPath, defaultEditor string) *exec.Cmd {
	if defaultEditor != "" {
		if runtime.GOOS == "windows" {
			cmd := exec.Command("cmd", "/c", "start", "/wait", "", defaultEditor, localPath)
			utils.SetHideWindow(cmd)
			return cmd
		}
		return exec.Command(defaultEditor, localPath)
	}

	if runtime.GOOS == "windows" {
		return exec.Command("notepad.exe", localPath)
	}
	if runtime.GOOS == "darwin" {
		return exec.Command("open", "-t", localPath)
	}

	editors := []string{"xdg-open", "gnome-text-editor", "kwrite", "gedit", "nano"}
	for _, e := range editors {
		if _, err := exec.LookPath(e); err == nil {
			return exec.Command(e, localPath)
		}
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

	return conn.SFTP.Getwd()
}
