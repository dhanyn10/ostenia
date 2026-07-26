package service

import (
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

func secureEnv() []string {
	var cleanEnv []string
	for _, envVar := range os.Environ() {
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

var wslRootPath = func(distro string) string {
	if RuntimeGOOS == "windows" {
		return `\\wsl.localhost\` + distro
	}
	return "/tmp/wsl/" + distro
}

func toWSLPath(wslRoot, remotePath string) string {
	cleaned := path.Clean(filepath.ToSlash(remotePath))
	if cleaned == "." || cleaned == "/" {
		return wslRoot
	}
	trimmed := strings.TrimPrefix(cleaned, "/")
	return filepath.Join(wslRoot, filepath.FromSlash(trimmed))
}

func decodeUTF16LE(b []byte) (string, error) {
	if len(b) < 2 {
		return string(b), nil
	}
	// Check for BOM (0xFF, 0xFE)
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

type WSLClient struct {
	Distro string
	User   string
}

func NewWSLClient(distro, user string) *WSLClient {
	return &WSLClient{Distro: distro, User: user}
}

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

func (c *WSLClient) NewSftp(opts ...sftp.ClientOption) (interfaces.SFTPClient, error) {
	root := wslRootPath(c.Distro)
	if RuntimeGOOS != "windows" {
		_ = os.MkdirAll(root, 0755)
	}
	return &WSLSFTPClient{
		Distro: c.Distro,
		Root:   root,
	}, nil
}

func (c *WSLClient) Close() error {
	return nil
}

type WSLSession struct {
	cmd    *exec.Cmd
	stdout io.Reader
	stdin  io.WriteCloser
	pipeW  io.WriteCloser
	distro string
	user   string
}

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

func (s *WSLSession) RequestPty(term string, h, w int, modes ssh.TerminalModes) error {
	return nil
}

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

func (s *WSLSession) WindowChange(h, w int) error {
	return nil
}

func (s *WSLSession) Close() error {
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	if s.pipeW != nil {
		_ = s.pipeW.Close()
	}
	return nil
}

type WSLSFTPClient struct {
	Distro string
	Root   string
}

func (w *WSLSFTPClient) ReadDir(p string) ([]os.FileInfo, error) {
	localPath := toWSLPath(w.Root, p)
	f, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdir(-1)
}

func (w *WSLSFTPClient) Stat(p string) (os.FileInfo, error) {
	localPath := toWSLPath(w.Root, p)
	return os.Stat(localPath)
}

func (w *WSLSFTPClient) RemoveAll(p string) error {
	localPath := toWSLPath(w.Root, p)
	return os.RemoveAll(localPath)
}

func (w *WSLSFTPClient) Remove(p string) error {
	localPath := toWSLPath(w.Root, p)
	return os.Remove(localPath)
}

func (w *WSLSFTPClient) Rename(oldpath, newpath string) error {
	localOld := toWSLPath(w.Root, oldpath)
	localNew := toWSLPath(w.Root, newpath)
	return os.Rename(localOld, localNew)
}

func (w *WSLSFTPClient) Mkdir(p string) error {
	localPath := toWSLPath(w.Root, p)
	return os.Mkdir(localPath, 0755)
}

func (w *WSLSFTPClient) Open(p string) (interfaces.SFTPFile, error) {
	localPath := toWSLPath(w.Root, p)
	return os.Open(localPath)
}

func (w *WSLSFTPClient) Create(p string) (interfaces.SFTPFile, error) {
	localPath := toWSLPath(w.Root, p)
	return os.Create(localPath)
}

func (w *WSLSFTPClient) Getwd() (string, error) {
	return "/", nil
}

func (w *WSLSFTPClient) Close() error {
	return nil
}
