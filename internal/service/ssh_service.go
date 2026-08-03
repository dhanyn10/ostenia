// Package service provides system-level service management, background orchestrations,
// symlink maintenance, SSH connections, and WSL client execution integrations for Ostenia.
package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"ostenia/internal/backend/interfaces"
	"ostenia/internal/config"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// wslPrefix is the scheme identifier used to detect local WSL quick-select connection hosts.
const wslPrefix = "wsl://"

// RemoteFile is an alias pointing to the backend's standard remote file description struct.
type RemoteFile = interfaces.RemoteFile

// SSHConnection wraps the complete active network connection state for a target session,
// including its SSH client, SFTP helper, interactive terminal stdin/stdout channels, and cancellable context.
type SSHConnection struct {
	SessionID string                // Unique identifier of the SSH session
	Client    interfaces.SSHClient  // SSH client wrapper (goph or custom WSL client)
	SFTP      interfaces.SFTPClient // SFTP client wrapper for file operations
	Shell     io.WriteCloser        // Stdin pipe to send interactive commands to the shell
	PTY       interfaces.SSHSession // Interactive terminal session (with PTY requested)
	Cancel    context.CancelFunc    // Function to cancel context and teardown this session
	LastSync  time.Time             // Timestamp of the last interactive synchronisation
	IsWSL     bool                  // Boolean flag indicating if this is a WSL loopback connection
}

// SSHManager coordinates active SSH and WSL interactive terminal and SFTP connections.
// It maintains a thread-safe registry of sessions, handles dialing/authentication,
// and manages dimensions/resizing of active virtual terminals (PTY).
type SSHManager struct {
	connections map[string]*SSHConnection // Registry mapping Session ID to active SSH connection
	mu          sync.RWMutex              // Read-Write Mutex to protect concurrent access to the registry
	dialer      SSHDialer                 // Extensible dialer function for mocking during tests
	runtime     interfaces.Runtime        // Wails runtime instance to emit output/disconnect events to the frontend
}

// SSHDialer is a function type representing standard SSH dial connections via goph.
type SSHDialer func(config *goph.Config) (interfaces.SSHClient, error)

// DefaultSSHDialer is the standard production implementation of SSHDialer using goph.NewConn.
var DefaultSSHDialer SSHDialer = func(config *goph.Config) (interfaces.SSHClient, error) {
	client, err := goph.NewConn(config)
	if err != nil {
		return nil, err
	}
	return &gophSSHClient{client}, nil
}

// NewSSHManager initializes and returns a new SSHManager instance.
func NewSSHManager() *SSHManager {
	return &SSHManager{
		connections: make(map[string]*SSHConnection),
		dialer:      DefaultSSHDialer,
	}
}

// SetRuntime registers the Wails runtime context within the SSH manager to enable frontend event emissions.
func (m *SSHManager) SetRuntime(r interfaces.Runtime) {
	m.runtime = r
}

// GetSessions retrieves the list of saved SSH sessions from the persistent global configuration.
func (m *SSHManager) GetSessions() ([]config.SSHSession, error) {
	return config.LoadSSHSessions()
}

// SaveSessions writes the complete list of SSH sessions back to the persistent global configuration.
func (m *SSHManager) SaveSessions(sessions []config.SSHSession) error {
	return config.SaveSSHSessions(sessions)
}

// dialSSHWithRetries establishes a connection with configured timeouts and retries.
func (m *SSHManager) dialSSHWithRetries(ctx context.Context, session config.SSHSession, auth goph.Auth) (interfaces.SSHClient, error) {
	timeout := 10 * time.Second
	if session.MaxTimeout > 0 {
		timeout = time.Duration(session.MaxTimeout) * time.Second
	}

	maxRetries := 3
	if session.MaxRetries > 0 {
		maxRetries = session.MaxRetries
	}

	var client interfaces.SSHClient
	var err error

	for i := 0; i < maxRetries; i++ {
		client, err = m.dialer(&goph.Config{
			User:     session.User,
			Addr:     session.Host,
			Port:     uint(session.Port),
			Auth:     auth,
			Timeout:  timeout,
			Callback: m.getHostKeyCallback(),
		})
		if err == nil {
			return client, nil
		}
		fmt.Printf("[SSH] Connection attempt %d/%d failed to %s: %v\n", i+1, maxRetries, session.Host, err)
		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(1 * time.Second):
			}
		}
	}
	return nil, err
}

// Connect establishes a connection to a remote SSH server or initiates a local WSL distribution terminal.
// It sets up the corresponding SFTP file explorer and triggers background terminal initialization.
func (m *SSHManager) Connect(ctx context.Context, session config.SSHSession) error {
	// Set UTF-8 environment variables for the current system process to avoid encoding mismatches.
	os.Setenv("LANG", "en_US.UTF-8")
	os.Setenv("LC_ALL", "en_US.UTF-8")

	m.mu.Lock()

	// If connection already exists, return early without reconnecting.
	if _, ok := m.connections[session.ID]; ok {
		m.mu.Unlock()
		return nil
	}

	fmt.Printf("[SSH] Connecting to %s@%s:%d...\n", session.User, session.Host, session.Port)

	var client interfaces.SSHClient
	var err error

	// WSL local connection detected via custom 'wsl://' host prefix.
	if strings.HasPrefix(session.Host, wslPrefix) {
		distro := strings.TrimPrefix(session.Host, wslPrefix)
		client = NewWSLClient(distro, session.User)
	} else {
		// Standard remote SSH connection. Retrieve authentication details first.
		var auth goph.Auth
		auth, err = m.getAuth(session)
		if err != nil {
			fmt.Printf("[SSH] Authentication retrieval failed: %v\n", err)
			m.mu.Unlock()
			return err
		}

		client, err = m.dialSSHWithRetries(ctx, session, auth)
		if err != nil {
			m.mu.Unlock()
			return err
		}
	}

	// Initialize SFTP Client for back-channel file exploration & transfers.
	sftpClient, err := client.NewSftp()
	if err != nil {
		fmt.Printf("[SSH] SFTP initialization failed: %v\n", err)
		client.Close()
		m.mu.Unlock()
		return err
	}

	isWSL := strings.HasPrefix(session.Host, wslPrefix)
	cCtx, cancel := context.WithCancel(ctx)
	conn := &SSHConnection{
		SessionID: session.ID,
		Client:    client,
		SFTP:      sftpClient,
		Cancel:    cancel,
		IsWSL:     isWSL,
	}

	m.connections[session.ID] = conn
	m.mu.Unlock()

	fmt.Printf("[SSH] Connected successfully to %s@%s:%d. Initializing terminal session.\n", session.User, session.Host, session.Port)

	// Spin up the interactive PTY terminal stream in the background.
	m.startTerminal(cCtx, conn)

	return nil
}

// SendInput writes raw keyboard / text input from the frontend terminal pane directly to the shell stdin.
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

// Disconnect cleans up and closes active shell processes, SFTP sockets, and root SSH connections.
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

// getAuth builds standard password-based, agent-based, or SSH key authentication setups.
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

// getHostKeyCallback verifies remote host keys, automatically writing unknown hosts to 'known_hosts'.
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

// GetWSLDistros executes a silent query requesting active Windows WSL distributions,
// decoding output from raw UTF-16LE.
func (m *SSHManager) GetWSLDistros() ([]string, error) {
	if RuntimeGOOS != "windows" {
		return []string{}, nil
	}

	cmd := wslCommand("", "-l", "-q")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	decoded, err := decodeUTF16LE(output)
	if err != nil {
		decoded = string(output)
	}

	var distros []string
	lines := strings.Split(decoded, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			distros = append(distros, line)
		}
	}
	return distros, nil
}

// ============================================================================
// Interfaces & Wrappers
// ============================================================================

// gophSSHClient wraps an active *goph.Client instance to fulfill the standard SSHClient interface.
type gophSSHClient struct {
	client *goph.Client
}

// NewSession starts and configures a new interactive shell session on top of the client.
func (c *gophSSHClient) NewSession() (interfaces.SSHSession, error) {
	s, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	return &sshSessionWrapper{s}, nil
}

// NewSftp hooks up a secure FTP transfer socket on top of the connection channel.
func (c *gophSSHClient) NewSftp(opts ...sftp.ClientOption) (interfaces.SFTPClient, error) {
	s, err := c.client.NewSftp(opts...)
	if err != nil {
		return nil, err
	}
	return &sftpClientWrapper{s}, nil
}

// Close gracefully closes the client network connection.
func (c *gophSSHClient) Close() error {
	return c.client.Close()
}

// sshSessionWrapper wraps a standard *ssh.Session instance to fulfill the standard SSHSession interface.
type sshSessionWrapper struct {
	session *ssh.Session
}

func (sw *sshSessionWrapper) StdoutPipe() (io.Reader, error)     { return sw.session.StdoutPipe() }
func (sw *sshSessionWrapper) StdinPipe() (io.WriteCloser, error) { return sw.session.StdinPipe() }
func (sw *sshSessionWrapper) RequestPty(term string, h, w int, modes ssh.TerminalModes) error {
	return sw.session.RequestPty(term, h, w, modes)
}
func (sw *sshSessionWrapper) Shell() error                { return sw.session.Shell() }
func (sw *sshSessionWrapper) Run(cmd string) error        { return sw.session.Run(cmd) }
func (sw *sshSessionWrapper) WindowChange(h, w int) error { return sw.session.WindowChange(h, w) }
func (sw *sshSessionWrapper) Close() error                { return sw.session.Close() }

// sftpClientWrapper wraps an active *sftp.Client instance to fulfill the standard SFTPClient interface.
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
