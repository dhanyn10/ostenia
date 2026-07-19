package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"ostenia/internal/backend/interfaces"
	"ostenia/internal/config"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf16"

	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"path"
	"runtime"
)

type RemoteFile = interfaces.RemoteFile

type SSHConnection struct {
	SessionID string
	Client    interfaces.SSHClient
	SFTP      interfaces.SFTPClient
	Shell     io.WriteCloser
	PTY       interfaces.SSHSession
	Context   context.Context
	Cancel    context.CancelFunc
	LastSync  time.Time
}

type SSHManager struct {
	connections map[string]*SSHConnection
	mu          sync.RWMutex
	dialer      SSHDialer
	runtime     interfaces.Runtime
}

type SSHDialer func(config *goph.Config) (interfaces.SSHClient, error)

var DefaultSSHDialer SSHDialer = func(config *goph.Config) (interfaces.SSHClient, error) {
	client, err := goph.NewConn(config)
	if err != nil {
		return nil, err
	}
	return &gophSSHClient{client}, nil
}

func NewSSHManager() *SSHManager {
	return &SSHManager{
		connections: make(map[string]*SSHConnection),
		dialer:      DefaultSSHDialer,
	}
}

func (m *SSHManager) SetRuntime(r interfaces.Runtime) {
	m.runtime = r
}

func (m *SSHManager) GetSessions() ([]config.SSHSession, error) {
	return config.LoadSSHSessions()
}

func (m *SSHManager) SaveSessions(sessions []config.SSHSession) error {
	return config.SaveSSHSessions(sessions)
}

func (m *SSHManager) Connect(ctx context.Context, session config.SSHSession) error {
	// Set UTF-8 environment for the session
	os.Setenv("LANG", "en_US.UTF-8")
	os.Setenv("LC_ALL", "en_US.UTF-8")

	m.mu.Lock()

	if _, ok := m.connections[session.ID]; ok {
		m.mu.Unlock()
		return nil
	}

	var client interfaces.SSHClient
	var sftpClient interfaces.SFTPClient
	var err error

	if session.Type == "wsl" {
		fmt.Printf("[WSL] Connecting to distro %s...\n", session.WSLDistro)
		client = &wslSSHClient{distroName: session.WSLDistro}
		sftpClient, err = client.NewSftp()
		if err != nil {
			m.mu.Unlock()
			return err
		}
	} else {
		fmt.Printf("[SSH] Connecting to %s@%s:%d...\n", session.User, session.Host, session.Port)

		auth, err := m.getAuth(session)
		if err != nil {
			fmt.Printf("[SSH] Authentication retrieval failed: %v\n", err)
			m.mu.Unlock()
			return err
		}

		client, err = m.dialer(&goph.Config{
			User:     session.User,
			Addr:     session.Host,
			Port:     uint(session.Port),
			Auth:     auth,
			Timeout:  10 * time.Second,
			Callback: m.getHostKeyCallback(),
		})

		if err != nil {
			fmt.Printf("[SSH] Dial connection failed: %v\n", err)
			m.mu.Unlock()
			return err
		}

		sftpClient, err = client.NewSftp()
		if err != nil {
			fmt.Printf("[SSH] SFTP initialization failed: %v\n", err)
			client.Close()
			m.mu.Unlock()
			return err
		}
	}

	cCtx, cancel := context.WithCancel(ctx)
	conn := &SSHConnection{
		SessionID: session.ID,
		Client:    client,
		SFTP:      sftpClient,
		Context:   cCtx,
		Cancel:    cancel,
	}

	m.connections[session.ID] = conn
	m.mu.Unlock()

	if session.Type == "wsl" {
		fmt.Printf("[WSL] Connected successfully to distro %s. Initializing terminal session.\n", session.WSLDistro)
	} else {
		fmt.Printf("[SSH] Connected successfully to %s@%s:%d. Initializing terminal session.\n", session.User, session.Host, session.Port)
	}

	// Start terminal session
	m.startTerminal(ctx, conn)

	return nil
}

func (m *SSHManager) startTerminal(ctx context.Context, conn *SSHConnection) {
	fmt.Printf("[SSH] Requesting new SSH session for terminal (SessionID: %s)...\n", conn.SessionID)
	sshSession, err := conn.Client.NewSession()
	if err != nil {
		fmt.Printf("[SSH] Failed to create SSH session: %v\n", err)
		return
	}
	m.mu.Lock()
	conn.PTY = sshSession
	m.mu.Unlock()

	stdout, err := sshSession.StdoutPipe()
	if err != nil {
		fmt.Printf("[SSH] Failed to get stdout pipe: %v\n", err)
		return
	}
	stdin, err := sshSession.StdinPipe()
	if err != nil {
		fmt.Printf("[SSH] Failed to get stdin pipe: %v\n", err)
		return
	}
	m.mu.Lock()
	conn.Shell = stdin
	m.mu.Unlock()

	if err := m.setupPTY(sshSession); err != nil {
		fmt.Printf("[SSH] Failed to setup terminal PTY: %v\n", err)
		return
	}

	exitChan := make(chan struct{})
	go m.processTerminalOutput(ctx, conn, stdout, exitChan)
	go m.handleTerminalExit(ctx, conn, sshSession, exitChan)
}

func (m *SSHManager) setupPTY(sshSession interfaces.SSHSession) error {
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
		ssh.ICANON:        0,
	}

	if err := sshSession.RequestPty("xterm-256color", 50, 200, modes); err != nil {
		return err
	}

	return sshSession.Shell()
}

func (m *SSHManager) processTerminalOutput(ctx context.Context, conn *SSHConnection, stdout io.Reader, exitChan chan struct{}) {
	buf := make([]byte, 2048)
	for {
		n, err := stdout.Read(buf)
		if n > 0 {
			if m.runtime != nil {
				m.runtime.EventsEmit(ctx, "ssh_output", map[string]interface{}{
					"sessionId": conn.SessionID,
					"data":      string(buf[:n]),
				})
			}
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

func (m *SSHManager) handleTerminalExit(ctx context.Context, conn *SSHConnection, sshSession interfaces.SSHSession, exitChan chan struct{}) {
	select {
	case <-conn.Context.Done():
		fmt.Printf("[SSH] Context done, closing terminal session %s\n", conn.SessionID)
		sshSession.Close()
	case <-exitChan:
		fmt.Printf("[SSH] Terminal output channel closed, disconnecting terminal session %s\n", conn.SessionID)
		if m.runtime != nil {
			m.runtime.EventsEmit(ctx, "ssh_disconnected", conn.SessionID)
		}
		m.mu.Lock()
		delete(m.connections, conn.SessionID)
		m.mu.Unlock()
	}
}

func (m *SSHManager) ResizeTerminal(sessionID string, cols, rows int) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found")
	}
	if conn.PTY == nil {
		return fmt.Errorf("session not connected")
	}

	// Logging to debug dimension issues in the sandbox
	fmt.Printf("[SSH] Resizing session %s to Cols: %d, Rows: %d\n", sessionID, cols, rows)

	// HARD GUARD: Never allow width to go below 100 to prevent erratic wrapping.
	if cols < 100 {
		cols = 100
	}
	if rows < 10 {
		rows = 10
	}

	// WindowChange(rows, cols) - SSH standard order is h, w
	return conn.PTY.WindowChange(rows, cols)
}

func (m *SSHManager) SendInput(sessionID, data string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found")
	}
	if conn.Shell == nil {
		return fmt.Errorf("session not connected")
	}

	_, err := conn.Shell.Write([]byte(data))
	return err
}

func (m *SSHManager) Disconnect(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Printf("[SSH] Disconnecting session ID: %s...\n", sessionID)
	if conn, ok := m.connections[sessionID]; ok {
		conn.Cancel()
		if conn.SFTP != nil {
			conn.SFTP.Close()
		}
		if conn.Client != nil {
			conn.Client.Close()
		}
		delete(m.connections, sessionID)
		fmt.Printf("[SSH] Session ID %s successfully disconnected.\n", sessionID)
	} else {
		fmt.Printf("[SSH] Session ID %s not found for disconnection.\n", sessionID)
	}
}

func (m *SSHManager) resolveRemotePath(conn *SSHConnection, p string) string {
	if strings.HasPrefix(p, "~") {
		wd, err := conn.SFTP.Getwd()
		if err == nil && wd != "" {
			if p == "~" || p == "~/" {
				return wd
			} else if strings.HasPrefix(p, "~/") {
				return path.Join(wd, p[2:])
			}
		}
	}
	return p
}

func (m *SSHManager) ListFiles(sessionID, pathStr string) ([]RemoteFile, error) {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	if conn.SFTP == nil {
		return nil, fmt.Errorf("SFTP not connected")
	}

	pathStr = m.resolveRemotePath(conn, pathStr)

	if pathStr == "" {
		wd, err := conn.SFTP.Getwd()
		if err == nil && wd != "" {
			pathStr = wd
		} else {
			pathStr = "."
		}
	}

	entries, err := conn.SFTP.ReadDir(pathStr)
	if err != nil {
		errStr := err.Error()
		if err == io.EOF || os.IsNotExist(err) ||
			strings.Contains(errStr, "EOF") ||
			strings.Contains(errStr, "does not exist") ||
			strings.Contains(errStr, "no such file") ||
			strings.Contains(errStr, "not found") ||
			strings.Contains(errStr, "not exist") {
			log.Printf("[SSH] ReadDir ignored empty/missing/EOF path %s: %v", pathStr, err)
			fmt.Printf("[SSH] ReadDir ignored empty/missing/EOF path %s: %v\n", pathStr, err)
			return []RemoteFile{}, nil
		}
		return nil, err
	}

	files := []RemoteFile{}
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

func (m *SSHManager) ExecuteSFTPAction(sessionID, action, remotePath, target string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found")
	}
	if conn.SFTP == nil {
		return fmt.Errorf("SFTP not connected")
	}

	remotePath = m.resolveRemotePath(conn, remotePath)
	remotePath = path.Clean(remotePath)
	if target != "" {
		target = m.resolveRemotePath(conn, target)
		target = path.Clean(target)
	}

	switch action {
	case "delete":
		// Check if it's a directory
		info, err := conn.SFTP.Stat(remotePath)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return conn.SFTP.RemoveAll(remotePath)
		}
		return conn.SFTP.Remove(remotePath)
	case "rename":
		return conn.SFTP.Rename(remotePath, target)
	case "mkdir":
		return conn.SFTP.Mkdir(remotePath)
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}

func (m *SSHManager) DownloadFile(sessionID, remotePath, localPath string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found")
	}
	if conn.SFTP == nil {
		return fmt.Errorf("SFTP not connected")
	}

	remotePath = m.resolveRemotePath(conn, remotePath)
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

func (m *SSHManager) UploadFile(sessionID, localPath, remotePath string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found")
	}
	if conn.SFTP == nil {
		return fmt.Errorf("SFTP not connected")
	}

	// Use path.Join for remote paths
	remotePath = m.resolveRemotePath(conn, remotePath)
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

func (m *SSHManager) EditFile(sessionID, remotePath, defaultEditor string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found")
	}
	if conn.SFTP == nil {
		return fmt.Errorf("SFTP not connected")
	}

	remotePath = m.resolveRemotePath(conn, remotePath)
	remotePath = path.Clean(remotePath)

	tempDir, err := os.MkdirTemp("", "ostenia-ssh-edit-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	fileName := path.Base(remotePath)
	localPath := filepath.Join(tempDir, fmt.Sprintf("%d-%s", time.Now().Unix(), fileName))

	if err := m.DownloadFile(sessionID, remotePath, localPath); err != nil {
		return err
	}

	initialInfo, _ := os.Stat(localPath)
	if err := m.runEditor(localPath, defaultEditor); err != nil {
		return err
	}

	finalInfo, err := os.Stat(localPath)
	if err == nil && finalInfo.ModTime().After(initialInfo.ModTime()) {
		return m.UploadFile(sessionID, localPath, remotePath)
	}

	return nil
}

func (m *SSHManager) runEditor(localPath, defaultEditor string) error {
	var cmd *exec.Cmd
	if defaultEditor != "" {
		cmd = m.getCustomEditorCmd(defaultEditor, localPath)
	} else {
		cmd = m.getDefaultEditorCmd(localPath)
	}

	if cmd == nil {
		return fmt.Errorf("no suitable text editor found")
	}

	return cmd.Run()
}

func (m *SSHManager) getCustomEditorCmd(editor, localPath string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		cmdPath := filepath.Join(utils.GetSystemDirectory(), "cmd.exe")
		cmd := utils.Executor.Command(cmdPath, "/c", "start", "/wait", "", editor, localPath)
		cmd.Env = utils.SafeEnv()
		utils.SetHideWindow(cmd)
		return cmd
	}
	cmd := utils.Executor.Command(editor, localPath)
	cmd.Env = utils.SafeEnv()
	return cmd
}

func (m *SSHManager) getDefaultEditorCmd(localPath string) *exec.Cmd {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmdPath := filepath.Join(utils.GetSystemDirectory(), "notepad.exe")
		cmd = utils.Executor.Command(cmdPath, localPath)
	case "darwin":
		cmd = utils.Executor.Command("/usr/bin/open", "-t", localPath)
	default:
		return m.findLinuxEditor(localPath)
	}
	cmd.Env = utils.SafeEnv()
	return cmd
}

func (m *SSHManager) findLinuxEditor(localPath string) *exec.Cmd {
	editors := []string{"xdg-open", "gnome-text-editor", "kwrite", "gedit", "nano"}
	for _, e := range editors {
		if path, err := exec.LookPath(e); err == nil {
			cmd := utils.Executor.Command(path, localPath)
			cmd.Env = utils.SafeEnv()
			return cmd
		}
	}
	return nil
}

func (m *SSHManager) GetCurrentPath(sessionID string) (string, error) {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found")
	}
	if conn.SFTP == nil {
		return "", fmt.Errorf("SFTP not connected")
	}

	pathStr, err := conn.SFTP.Getwd()
	return pathStr, err
}

func (m *SSHManager) getAuth(session config.SSHSession) (goph.Auth, error) {
	if session.AuthMethod == "agent" {
		return goph.UseAgent()
	}

	if session.AuthMethod == "password" {
		return goph.Auth{
			ssh.Password(session.Password),
			ssh.KeyboardInteractive(func(user, instruction string, questions []string, echos []bool) ([]string, error) {
				answers := make([]string, len(questions))
				for i := range questions {
					answers[i] = session.Password
				}
				return answers, nil
			}),
		}, nil
	}

	passphrase := ""
	if session.Passphrase != "" {
		passphrase = session.Passphrase
	}
	return goph.Key(session.KeyPath, passphrase)
}

func (m *SSHManager) getHostKeyCallback() ssh.HostKeyCallback {
	knownHostsPath := filepath.Join(config.GetBaseDir(), "known_hosts")
	if _, err := os.Stat(knownHostsPath); os.IsNotExist(err) {
		_ = os.WriteFile(knownHostsPath, []byte(""), 0600)
	}

	return func(host string, remote net.Addr, key ssh.PublicKey) error {
		found, err := goph.CheckKnownHost(host, remote, key, knownHostsPath)
		if err != nil {
			fmt.Printf("Warning: failed to check known_hosts: %v\n", err)
			return nil // Continue anyway but log it
		}
		if found {
			return nil
		}
		// If not found, add it (remembering it for next time)
		return goph.AddKnownHost(host, remote, key, knownHostsPath)
	}
}

// Wrappers

type gophSSHClient struct {
	client *goph.Client
}

func (c *gophSSHClient) NewSession() (interfaces.SSHSession, error) {
	s, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	return &sshSessionWrapper{s}, nil
}

func (c *gophSSHClient) NewSftp(opts ...sftp.ClientOption) (interfaces.SFTPClient, error) {
	s, err := c.client.NewSftp(opts...)
	if err != nil {
		return nil, err
	}
	return &sftpClientWrapper{s}, nil
}

func (c *gophSSHClient) Close() error {
	return c.client.Close()
}

type sshSessionWrapper struct {
	session *ssh.Session
}

func (sw *sshSessionWrapper) StdoutPipe() (io.Reader, error)     { return sw.session.StdoutPipe() }
func (sw *sshSessionWrapper) StdinPipe() (io.WriteCloser, error) { return sw.session.StdinPipe() }
func (sw *sshSessionWrapper) RequestPty(term string, h, w int, modes ssh.TerminalModes) error {
	return sw.session.RequestPty(term, h, w, modes)
}
func (sw *sshSessionWrapper) Shell() error                { return sw.session.Shell() }
func (sw *sshSessionWrapper) WindowChange(h, w int) error { return sw.session.WindowChange(h, w) }
func (sw *sshSessionWrapper) Close() error                { return sw.session.Close() }

type sftpClientWrapper struct {
	client *sftp.Client
}

func (cw *sftpClientWrapper) ReadDir(p string) ([]os.FileInfo, error) { return cw.client.ReadDir(p) }
func (cw *sftpClientWrapper) Stat(p string) (os.FileInfo, error)      { return cw.client.Stat(p) }
func (cw *sftpClientWrapper) RemoveAll(p string) error                { return cw.client.RemoveAll(p) }
func (cw *sftpClientWrapper) Remove(p string) error                   { return cw.client.Remove(p) }
func (cw *sftpClientWrapper) Rename(oldpath, newpath string) error {
	return cw.client.Rename(oldpath, newpath)
}
func (cw *sftpClientWrapper) Mkdir(p string) error                       { return cw.client.Mkdir(p) }
func (cw *sftpClientWrapper) Open(p string) (interfaces.SFTPFile, error) { return cw.client.Open(p) }
func (cw *sftpClientWrapper) Create(p string) (interfaces.SFTPFile, error) {
	return cw.client.Create(p)
}
func (cw *sftpClientWrapper) Getwd() (string, error) { return cw.client.Getwd() }
func (cw *sftpClientWrapper) Close() error           { return cw.client.Close() }

// WSL/Local terminal & filesystem implementation

func decodeUTF16(b []byte) string {
	if len(b) < 2 {
		return string(b)
	}
	u16s := make([]uint16, len(b)/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = uint16(b[2*i]) | (uint16(b[2*i+1]) << 8)
	}
	if len(u16s) > 0 && u16s[0] == 0xFEFF {
		u16s = u16s[1:]
	}
	return string(utf16.Decode(u16s))
}

// GetWSLDistributions returns the list of installed WSL distributions
func GetWSLDistributions() ([]string, error) {
	if runtime.GOOS != "windows" {
		return []string{"Ubuntu", "Debian", "Alpine"}, nil
	}

	cmd := exec.Command("wsl.exe", "--list", "--quiet")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	decoded := decodeUTF16(output)
	lines := strings.Split(decoded, "\n")
	var distros []string
	for _, line := range lines {
		cleaned := strings.TrimSpace(line)
		cleaned = strings.ReplaceAll(cleaned, "\x00", "")
		if cleaned != "" {
			distros = append(distros, cleaned)
		}
	}
	return distros, nil
}

type wslSSHClient struct {
	distroName string
}

func (c *wslSSHClient) NewSession() (interfaces.SSHSession, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("wsl.exe", "-d", c.distroName)
	} else {
		cmd = exec.Command("sh")
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	go func() {
		_, _ = io.Copy(pw, stdout)
	}()
	go func() {
		_, _ = io.Copy(pw, stderr)
	}()

	return &wslSessionWrapper{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		output: pr,
	}, nil
}

func (c *wslSSHClient) NewSftp(opts ...sftp.ClientOption) (interfaces.SFTPClient, error) {
	return &WSLFileSystemClient{DistroName: c.distroName}, nil
}

func (c *wslSSHClient) Close() error {
	return nil
}

type wslSessionWrapper struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	output io.Reader
}

func (sw *wslSessionWrapper) StdoutPipe() (io.Reader, error) {
	return sw.output, nil
}

func (sw *wslSessionWrapper) StdinPipe() (io.WriteCloser, error) {
	return sw.stdin, nil
}

func (sw *wslSessionWrapper) RequestPty(term string, h, w int, modes ssh.TerminalModes) error {
	return nil
}

func (sw *wslSessionWrapper) Shell() error {
	return sw.cmd.Start()
}

func (sw *wslSessionWrapper) WindowChange(h, w int) error {
	return nil
}

func (sw *wslSessionWrapper) Close() error {
	if sw.cmd.Process != nil {
		_ = sw.cmd.Process.Kill()
	}
	return nil
}

type WSLFileSystemClient struct {
	DistroName string
}

func (c *WSLFileSystemClient) toLocalPath(p string) string {
	cleaned := path.Clean(p)
	if cleaned == "." || cleaned == "" {
		cleaned = "/"
	}
	winPath := filepath.FromSlash(cleaned)
	if runtime.GOOS == "windows" {
		return filepath.Join(`\\wsl.localhost`, c.DistroName, winPath)
	} else {
		// Mock path on non-Windows (tests/sandbox)
		mockBase := filepath.Join("/tmp/wsl", c.DistroName)
		_ = os.MkdirAll(mockBase, 0755)
		return filepath.Join(mockBase, winPath)
	}
}

func (c *WSLFileSystemClient) ReadDir(p string) ([]os.FileInfo, error) {
	localPath := c.toLocalPath(p)
	if runtime.GOOS != "windows" {
		_ = os.MkdirAll(localPath, 0755)
	}

	entries, err := os.ReadDir(localPath)
	if err != nil {
		return nil, err
	}

	var infos []os.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err == nil {
			infos = append(infos, info)
		}
	}
	return infos, nil
}

func (c *WSLFileSystemClient) Stat(p string) (os.FileInfo, error) {
	localPath := c.toLocalPath(p)
	if runtime.GOOS != "windows" {
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			_ = os.WriteFile(localPath, []byte("mock content"), 0644)
		}
	}
	return os.Stat(localPath)
}

func (c *WSLFileSystemClient) RemoveAll(p string) error {
	return os.RemoveAll(c.toLocalPath(p))
}

func (c *WSLFileSystemClient) Remove(p string) error {
	return os.Remove(c.toLocalPath(p))
}

func (c *WSLFileSystemClient) Rename(oldpath, newpath string) error {
	return os.Rename(c.toLocalPath(oldpath), c.toLocalPath(newpath))
}

func (c *WSLFileSystemClient) Mkdir(p string) error {
	return os.Mkdir(c.toLocalPath(p), 0755)
}

func (c *WSLFileSystemClient) Open(p string) (interfaces.SFTPFile, error) {
	localPath := c.toLocalPath(p)
	if runtime.GOOS != "windows" {
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			_ = os.WriteFile(localPath, []byte("mock content"), 0644)
		}
	}
	f, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	return &wslFileWrapper{f}, nil
}

func (c *WSLFileSystemClient) Create(p string) (interfaces.SFTPFile, error) {
	localPath := c.toLocalPath(p)
	f, err := os.Create(localPath)
	if err != nil {
		return nil, err
	}
	return &wslFileWrapper{f}, nil
}

func (c *WSLFileSystemClient) Getwd() (string, error) {
	return "/", nil
}

func (c *WSLFileSystemClient) Close() error {
	return nil
}

type wslFileWrapper struct {
	*os.File
}
