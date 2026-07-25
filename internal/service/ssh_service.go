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

	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"path"
	"runtime"
	"unicode/utf16"
)

const wslPrefix = "wsl://"

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
	IsWSL     bool
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

	fmt.Printf("[SSH] Connecting to %s@%s:%d...\n", session.User, session.Host, session.Port)

	var client interfaces.SSHClient
	var err error

	if strings.HasPrefix(session.Host, wslPrefix) {
		distro := strings.TrimPrefix(session.Host, wslPrefix)
		client = NewWSLClient(distro, session.User)
	} else {
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
	}

	if err != nil {
		fmt.Printf("[SSH] Dial connection failed: %v\n", err)
		m.mu.Unlock()
		return err
	}

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
		Context:   cCtx,
		Cancel:    cancel,
		IsWSL:     isWSL,
	}

	m.connections[session.ID] = conn
	m.mu.Unlock()

	fmt.Printf("[SSH] Connected successfully to %s@%s:%d. Initializing terminal session.\n", session.User, session.Host, session.Port)

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

	if conn.IsWSL && conn.Shell != nil {
		_, _ = conn.Shell.Write([]byte("\n"))
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

func (m *SSHManager) GetResourceUsage(sessionID string) (interfaces.ResourceUsage, error) {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return interfaces.ResourceUsage{}, fmt.Errorf("session not found")
	}

	sess, err := conn.Client.NewSession()
	if err != nil {
		return interfaces.ResourceUsage{}, err
	}
	defer sess.Close()

	stdout, err := sess.StdoutPipe()
	if err != nil {
		return interfaces.ResourceUsage{}, err
	}

	command := `cat /proc/stat; sleep 0.1; cat /proc/stat; cat /proc/meminfo; df -m /`

	err = sess.Run(command)
	if err != nil {
		return interfaces.ResourceUsage{}, err
	}

	outputBytes, err := io.ReadAll(stdout)
	if err != nil {
		return interfaces.ResourceUsage{}, err
	}

	return parseResourceUsage(decodeMaybeUTF16(outputBytes)), nil
}

func decodeMaybeUTF16(b []byte) string {
	if len(b) < 2 {
		return string(b)
	}

	isUTF16 := false
	if b[0] == 0xFF && b[1] == 0xFE {
		isUTF16 = true
	} else if len(b) >= 4 && b[1] == 0x00 && b[3] == 0x00 {
		isUTF16 = true
	}

	if !isUTF16 {
		return string(b)
	}

	startIdx := 0
	if b[0] == 0xFF && b[1] == 0xFE {
		startIdx = 2
	}

	bytesToDecode := b[startIdx:]
	if len(bytesToDecode)%2 != 0 {
		bytesToDecode = bytesToDecode[:len(bytesToDecode)-1]
	}

	u16 := make([]uint16, len(bytesToDecode)/2)
	for i := 0; i < len(u16); i++ {
		u16[i] = uint16(bytesToDecode[2*i]) | (uint16(bytesToDecode[2*i+1]) << 8)
	}
	return string(utf16.Decode(u16))
}

func parseResourceUsage(output string) interfaces.ResourceUsage {
	var usage interfaces.ResourceUsage
	lines := strings.Split(output, "\n")

	var cpuTicks [][]float64
	var memTotal, memAvailable, memFree, buffers, cached float64
	var diskTotal, diskUsed float64

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)

		if strings.HasPrefix(line, "cpu ") && len(fields) >= 5 {
			var user, nice, system, idle, iowait, irq, softirq, steal float64
			fmt.Sscanf(fields[1], "%f", &user)
			fmt.Sscanf(fields[2], "%f", &nice)
			fmt.Sscanf(fields[3], "%f", &system)
			fmt.Sscanf(fields[4], "%f", &idle)
			if len(fields) >= 6 {
				fmt.Sscanf(fields[5], "%f", &iowait)
			}
			if len(fields) >= 7 {
				fmt.Sscanf(fields[6], "%f", &irq)
			}
			if len(fields) >= 8 {
				fmt.Sscanf(fields[7], "%f", &softirq)
			}
			if len(fields) >= 9 {
				fmt.Sscanf(fields[8], "%f", &steal)
			}

			total := user + nice + system + idle + iowait + irq + softirq + steal
			idleVal := idle + iowait
			cpuTicks = append(cpuTicks, []float64{total, idleVal})
		}

		if strings.HasPrefix(line, "MemTotal:") {
			fmt.Sscanf(line, "MemTotal: %f kB", &memTotal)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fmt.Sscanf(line, "MemAvailable: %f kB", &memAvailable)
		} else if strings.HasPrefix(line, "MemFree:") {
			fmt.Sscanf(line, "MemFree: %f kB", &memFree)
		} else if strings.HasPrefix(line, "Buffers:") {
			fmt.Sscanf(line, "Buffers: %f kB", &buffers)
		} else if strings.HasPrefix(line, "Cached:") {
			fmt.Sscanf(line, "Cached: %f kB", &cached)
		}

		if len(fields) >= 5 && fields[len(fields)-1] == "/" {
			if len(fields) >= 6 {
				var tot, usd float64
				fmt.Sscanf(fields[1], "%f", &tot)
				fmt.Sscanf(fields[2], "%f", &usd)
				if tot > 0 {
					diskTotal = tot
					diskUsed = usd
				}
			} else if len(fields) == 5 {
				var tot, usd float64
				fmt.Sscanf(fields[0], "%f", &tot)
				fmt.Sscanf(fields[1], "%f", &usd)
				if tot > 0 {
					diskTotal = tot
					diskUsed = usd
				}
			}
		}
	}

	if len(cpuTicks) >= 2 {
		diffTotal := cpuTicks[1][0] - cpuTicks[0][0]
		diffIdle := cpuTicks[1][1] - cpuTicks[0][1]
		if diffTotal > 0 {
			usage.CPU = ((diffTotal - diffIdle) / diffTotal) * 100
		}
	}

	if memTotal > 0 {
		usage.MemTotal = memTotal / 1024
		var used float64
		if memAvailable > 0 {
			used = memTotal - memAvailable
		} else {
			used = memTotal - memFree - buffers - cached
		}
		usage.MemUsed = used / 1024
		usage.Mem = (used / memTotal) * 100
	}

	if diskTotal > 0 {
		usage.DiskTotal = diskTotal
		usage.DiskUsed = diskUsed
		usage.Disk = (diskUsed / diskTotal) * 100
	}

	return usage
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
func (sw *sshSessionWrapper) Run(cmd string) error        { return sw.session.Run(cmd) }
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
