package service

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"ostenia/internal/backend/interfaces"
	plugins_utils "ostenia/internal/plugins/utils"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// secureEnv constructs a safe execution environment by filtering out unsafe variables
// and setting standard secure search path variables (PATH) for Windows and non-Windows hosts.
func secureEnv() []string {
	var cleanEnv []string
	for _, envVar := range os.Environ() {
		// Strip volatile or user-modifiable variables to ensure command isolation.
		if !strings.HasPrefix(strings.ToUpper(envVar), "PATH=") && !strings.HasPrefix(strings.ToUpper(envVar), "TERM=") {
			cleanEnv = append(cleanEnv, envVar)
		}
	}

	var safePath string
	if RuntimeGOOS == "windows" {
		systemRoot := os.Getenv("SystemRoot")
		if systemRoot == "" {
			systemRoot = `C:\Windows`
		}
		paths := []string{
			filepath.Join(systemRoot, "System32"),
			systemRoot,
			filepath.Join(systemRoot, "System32", "Wbem"),
			filepath.Join(systemRoot, "System32", "WindowsPowerShell", "v1.0"),
		}
		safePath = "PATH=" + strings.Join(paths, ";") // NOSONAR
	} else {
		// Restrict strictly to unwriteable system directories, excluding writable locations like /usr/local/bin
		safePath = "PATH=/usr/bin:/bin:/usr/sbin:/sbin" // NOSONAR
	}
	cleanEnv = append(cleanEnv, safePath) // NOSONAR
	cleanEnv = append(cleanEnv, "TERM=xterm-256color")
	return cleanEnv
}

// wslCommand builds and configures a wsl.exe process runner.
// Includes mock fallbacks for Unix/macOS environments to keep unit tests stable in CI/CD pipelines.
var wslCommand = func(distro string, args ...string) *exec.Cmd {
	var cmd *exec.Cmd
	if RuntimeGOOS == "windows" {
		cmdArgs := []string{"-d", distro}
		if len(args) == 0 {
			cmdArgs = append(cmdArgs, "bash", "--login", "-i")
		} else {
			cmdArgs = append(cmdArgs, args...)
		}
		cmd = exec.Command("wsl.exe", cmdArgs...) // NOSONAR
	} else {
		// Fallback/mock implementation for tests on non-Windows platforms
		if len(args) == 0 {
			cmd = exec.Command("sh", "-c", "echo 'WSL Shell Mock'; sleep 0.1") // NOSONAR
		} else {
			cmdArgs := append([]string{"-c"}, args...)
			cmd = exec.Command("sh", cmdArgs...) // NOSONAR
		}
	}
	cmd.Env = secureEnv()
	plugins_utils.SetHideWindow(cmd)
	return cmd
}

// wslRootPath maps the WSL virtual root path format to local directory trees depending on the host OS.
var wslRootPath = func(distro string) string {
	if RuntimeGOOS == "windows" {
		return `\\wsl.localhost\` + distro
	}
	return "/tmp/wsl/" + distro
}

// toWSLPath translates standard SFTP paths to physical folders mapped inside the host computer.
func toWSLPath(wslRoot, remotePath string) string {
	cleaned := path.Clean(filepath.ToSlash(remotePath))
	if cleaned == "." || cleaned == "/" {
		return wslRoot
	}
	trimmed := strings.TrimPrefix(cleaned, "/")
	return filepath.Join(wslRoot, filepath.FromSlash(trimmed))
}

// decodeUTF16LE converts binary arrays containing UTF-16 Little Endian encoded strings (as output by wsl.exe)
// into standard UTF-8 text representation.
func decodeUTF16LE(b []byte) (string, error) {
	if len(b) < 2 {
		return string(b), nil
	}
	// Detect and skip UTF-16 Byte Order Mark (BOM) [0xFF, 0xFE] if present
	hasBOM := b[0] == 0xFF && b[1] == 0xFE
	startIdx := 0
	if hasBOM {
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
	return string(utf16.Decode(u16)), nil
}

// WSLClient manages integration with local WSL distributions by mocking SSH / SFTP APIs locally.
type WSLClient struct {
	Distro string // Name of the target WSL distribution (e.g., Ubuntu, Debian)
	User   string // Execution user context (e.g., root, ubuntu)
}

// NewWSLClient creates a new WSLClient instance.
func NewWSLClient(distro, user string) *WSLClient {
	return &WSLClient{Distro: distro, User: user}
}

// NewSession starts and spawns a WSL interactive shell process.
func (c *WSLClient) NewSession() (interfaces.SSHSession, error) {
	var cmd *exec.Cmd
	if c.User != "" && c.User != "root" {
		cmd = wslCommand(c.Distro, "-u", c.User, "bash", "--login", "-i")
	} else {
		cmd = wslCommand(c.Distro)
	}
	return &WSLSession{
		cmd:    cmd,
		distro: c.Distro,
		user:   c.User,
	}, nil
}

// NewSftp initiates a mock SFTP handler pointing to the distro's virtual folder mapping.
func (c *WSLClient) NewSftp(opts ...sftp.ClientOption) (interfaces.SFTPClient, error) {
	root := wslRootPath(c.Distro)
	if RuntimeGOOS != "windows" {
		_ = os.MkdirAll(root, 0755)
	}
	return &WSLSFTPClient{
		Distro: c.Distro,
		User:   c.User,
		Root:   root,
	}, nil
}

// Close closes any active state inside the WSL client handler (no-op).
func (c *WSLClient) Close() error {
	return nil
}

// WSLSession wraps a running WSL shell subprocess, acting as an interactive terminal stream wrapper.
type WSLSession struct {
	cmd    *exec.Cmd      // Active command runner
	stdout io.Reader      // Read pipe for stdout redirection
	stdin  io.WriteCloser // Write pipe for stdin input feeding
	pipeW  io.WriteCloser // Helper stream writer to capture background outputs
	distro string         // Target WSL distro identifier
	user   string         // Specific Linux system execution user
}

// StdoutPipe sets up and hooks up stdout/stderr output pipes of the WSL subprocess.
func (s *WSLSession) StdoutPipe() (io.Reader, error) {
	if s.stdout != nil {
		return s.stdout, nil
	}
	pr, pw := io.Pipe()
	s.cmd.Stdout = pw
	s.cmd.Stderr = pw
	s.stdout = pr
	s.pipeW = pw
	return pr, nil
}

// StdinPipe retrieves the writable input stdin channel of the active WSL subprocess.
func (s *WSLSession) StdinPipe() (io.WriteCloser, error) {
	if s.stdin != nil {
		return s.stdin, nil
	}
	stdin, err := s.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	s.stdin = stdin
	return stdin, nil
}

// RequestPty mocks virtual terminal PTY requests (not required for direct subprocess streams).
func (s *WSLSession) RequestPty(term string, h, w int, modes ssh.TerminalModes) error {
	return nil
}

// Shell launches the underlying interactive bash subprocess command stream.
func (s *WSLSession) Shell() error {
	err := s.cmd.Start()
	if err != nil {
		return err
	}
	if s.pipeW != nil {
		go func() {
			_ = s.cmd.Wait()
			_ = s.pipeW.Close()
		}()
	}
	return nil
}

// Run executes a separate single WSL command and blocks until completion.
func (s *WSLSession) Run(cmd string) error {
	if cmd != "" {
		var newCmd *exec.Cmd
		if s.user != "" && s.user != "root" {
			newCmd = wslCommand(s.distro, "-u", s.user, "sh", "-c", cmd)
		} else {
			newCmd = wslCommand(s.distro, "sh", "-c", cmd)
		}
		if s.pipeW != nil {
			newCmd.Stdout = s.cmd.Stdout
			newCmd.Stderr = s.cmd.Stderr
		}
		s.cmd = newCmd
	}
	err := s.cmd.Start()
	if err != nil {
		return err
	}
	if s.pipeW != nil {
		go func() {
			_ = s.cmd.Wait()
			_ = s.pipeW.Close()
		}()
		return nil
	}
	return s.cmd.Wait()
}

// WindowChange mocks PTY resize events for Windows CMD/PowerShell subshells.
func (s *WSLSession) WindowChange(h, w int) error {
	return nil
}

// Close terminates and kills the underlying WSL subprocess.
func (s *WSLSession) Close() error {
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	if s.pipeW != nil {
		_ = s.pipeW.Close()
	}
	return nil
}

// WSLSFTPClient implements SFTP APIs directly on the local host to expose WSL container filesystem exploration.
type WSLSFTPClient struct {
	Distro string // Name of the target WSL distribution
	User   string // Execution user context
	Root   string // Local root mapping directory to query files from
}

// dummySFTPFile wraps fallback mocked empty files under permissions errors.
type dummySFTPFile struct{}

func (d *dummySFTPFile) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}
func (d *dummySFTPFile) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("write not supported on dummy sftp file: access denied")
}
func (d *dummySFTPFile) Close() error {
	return nil
}
func (d *dummySFTPFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}
func (d *dummySFTPFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, io.EOF
}
func (d *dummySFTPFile) WriteAt(p []byte, off int64) (n int, err error) {
	return 0, fmt.Errorf("write not supported on dummy sftp file: access denied")
}

// ReadDir returns the contents of a directory on the local WSL path.
func (w *WSLSFTPClient) ReadDir(p string) ([]os.FileInfo, error) {
	localPath := toWSLPath(w.Root, p)
	f, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdir(-1)
}

// Stat retrieves system details of a WSL file or directory.
func (w *WSLSFTPClient) Stat(p string) (os.FileInfo, error) {
	localPath := toWSLPath(w.Root, p)
	return os.Stat(localPath)
}

// RemoveAll recursively removes files and subdirectories.
func (w *WSLSFTPClient) RemoveAll(p string) error {
	localPath := toWSLPath(w.Root, p)
	err := os.RemoveAll(localPath)
	if err != nil && RuntimeGOOS == "windows" && os.IsPermission(err) {
		var cmd *exec.Cmd
		if w.User != "" && w.User != "root" {
			cmd = exec.Command("wsl.exe", "-d", w.Distro, "-u", w.User, "rm", "-rf", p)
		} else {
			cmd = exec.Command("wsl.exe", "-d", w.Distro, "rm", "-rf", p)
		}
		plugins_utils.SetHideWindow(cmd)
		return cmd.Run()
	}
	return err
}

// Remove deletes a single file in the WSL tree.
func (w *WSLSFTPClient) Remove(p string) error {
	localPath := toWSLPath(w.Root, p)
	err := os.Remove(localPath)
	if err != nil && RuntimeGOOS == "windows" && os.IsPermission(err) {
		var cmd *exec.Cmd
		if w.User != "" && w.User != "root" {
			cmd = exec.Command("wsl.exe", "-d", w.Distro, "-u", w.User, "rm", p)
		} else {
			cmd = exec.Command("wsl.exe", "-d", w.Distro, "rm", p)
		}
		plugins_utils.SetHideWindow(cmd)
		return cmd.Run()
	}
	return err
}

// Rename moves or renames a WSL file pathway.
func (w *WSLSFTPClient) Rename(oldpath, newpath string) error {
	localOld := toWSLPath(w.Root, oldpath)
	localNew := toWSLPath(w.Root, newpath)
	err := os.Rename(localOld, localNew)
	if err != nil && RuntimeGOOS == "windows" && os.IsPermission(err) {
		var cmd *exec.Cmd
		if w.User != "" && w.User != "root" {
			cmd = exec.Command("wsl.exe", "-d", w.Distro, "-u", w.User, "mv", oldpath, newpath)
		} else {
			cmd = exec.Command("wsl.exe", "-d", w.Distro, "mv", oldpath, newpath)
		}
		plugins_utils.SetHideWindow(cmd)
		return cmd.Run()
	}
	return err
}

// Mkdir creates a new folder pathway inside WSL.
func (w *WSLSFTPClient) Mkdir(p string) error {
	if RuntimeGOOS == "windows" {
		var cmd *exec.Cmd
		if w.User != "" && w.User != "root" {
			cmd = exec.Command("wsl.exe", "-d", w.Distro, "-u", w.User, "mkdir", "-p", p)
		} else {
			cmd = exec.Command("wsl.exe", "-d", w.Distro, "mkdir", "-p", p)
		}
		plugins_utils.SetHideWindow(cmd)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	localPath := toWSLPath(w.Root, p)
	return os.Mkdir(localPath, 0755)
}

// Open reads a specific file pathway from the WSL tree.
func (w *WSLSFTPClient) Open(p string) (interfaces.SFTPFile, error) {
	localPath := toWSLPath(w.Root, p)
	return os.Open(localPath)
}

// Create generates a new file pathway on the WSL tree.
func (w *WSLSFTPClient) Create(p string) (interfaces.SFTPFile, error) {
	if RuntimeGOOS == "windows" {
		var cmd *exec.Cmd
		if w.User != "" && w.User != "root" {
			cmd = exec.Command("wsl.exe", "-d", w.Distro, "-u", w.User, "touch", p)
		} else {
			cmd = exec.Command("wsl.exe", "-d", w.Distro, "touch", p)
		}
		plugins_utils.SetHideWindow(cmd)
		_ = cmd.Run()
	}
	localPath := toWSLPath(w.Root, p)
	f, err := os.Create(localPath)
	if err != nil {
		if RuntimeGOOS == "windows" && os.IsPermission(err) {
			if _, statErr := os.Stat(localPath); statErr == nil {
				return &dummySFTPFile{}, nil
			}
		}
		return nil, err
	}
	return f, nil
}

// Getwd returns the current active home/root directory of the mock SFTP client.
func (w *WSLSFTPClient) Getwd() (string, error) {
	if RuntimeGOOS == "windows" {
		userStr := "root"
		if w.User != "" {
			userStr = w.User
		}
		cmd := exec.Command("wsl.exe", "-d", w.Distro, "-u", userStr, "sh", "-c", "echo $HOME")
		plugins_utils.SetHideWindow(cmd)
		output, err := cmd.Output()
		if err == nil {
			trimmed := strings.TrimSpace(string(output))
			if trimmed != "" && strings.HasPrefix(trimmed, "/") {
				return trimmed, nil
			}
		}
	}
	return "/", nil
}

// Close gracefully closes the WSL SFTP client.
func (w *WSLSFTPClient) Close() error {
	return nil
}
