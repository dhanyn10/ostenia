package service

import (
	"io"
	"os"
	"os/exec"
	"ostenia/internal/backend/interfaces"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type WSLClient struct {
	distro string
}

func NewWSLClient(distro string) (*WSLClient, error) {
	return &WSLClient{distro: distro}, nil
}

func (c *WSLClient) NewSession() (interfaces.SSHSession, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("wsl.exe", "-d", c.distro)
	} else {
		cmd = exec.Command("sh")
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	cmd.Stderr = cmd.Stdout

	return &WSLSession{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}, nil
}

func (c *WSLClient) NewSftp(opts ...sftp.ClientOption) (interfaces.SFTPClient, error) {
	return &WSLSftpClient{distro: c.distro}, nil
}

func (c *WSLClient) Close() error {
	return nil
}

type WSLSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (s *WSLSession) StdoutPipe() (io.Reader, error) {
	return s.stdout, nil
}

func (s *WSLSession) StdinPipe() (io.WriteCloser, error) {
	return s.stdin, nil
}

func (s *WSLSession) RequestPty(term string, h, w int, modes ssh.TerminalModes) error {
	return nil
}

func (s *WSLSession) Shell() error {
	return s.cmd.Start()
}

func (s *WSLSession) WindowChange(h, w int) error {
	return nil
}

func (s *WSLSession) Close() error {
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	_ = s.stdout.Close()
	_ = s.stdin.Close()
	return nil
}

type WSLSftpClient struct {
	distro string
}

func (c *WSLSftpClient) mapPath(p string) string {
	if p == "" {
		p = "/"
	}
	p = filepath.ToSlash(p)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	var root string
	if runtime.GOOS == "windows" {
		root = "\\\\wsl.localhost\\" + c.distro
	} else {
		root = filepath.Join(os.TempDir(), "wsl", c.distro)
		_ = os.MkdirAll(root, 0755)
	}

	parts := strings.Split(p, "/")
	elem := []string{root}
	for _, part := range parts {
		if part != "" {
			elem = append(elem, part)
		}
	}
	return filepath.Join(elem...)
}

func (c *WSLSftpClient) ReadDir(p string) ([]os.FileInfo, error) {
	localPath := c.mapPath(p)
	entries, err := os.ReadDir(localPath)
	if err != nil {
		return nil, err
	}
	infos := make([]os.FileInfo, len(entries))
	for i, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		infos[i] = info
	}
	return infos, nil
}

func (c *WSLSftpClient) Stat(p string) (os.FileInfo, error) {
	localPath := c.mapPath(p)
	return os.Stat(localPath)
}

func (c *WSLSftpClient) RemoveAll(p string) error {
	localPath := c.mapPath(p)
	return os.RemoveAll(localPath)
}

func (c *WSLSftpClient) Remove(p string) error {
	localPath := c.mapPath(p)
	return os.Remove(localPath)
}

func (c *WSLSftpClient) Rename(oldpath, newpath string) error {
	localOld := c.mapPath(oldpath)
	localNew := c.mapPath(newpath)
	_ = os.MkdirAll(filepath.Dir(localNew), 0755)
	return os.Rename(localOld, localNew)
}

func (c *WSLSftpClient) Mkdir(p string) error {
	localPath := c.mapPath(p)
	return os.MkdirAll(localPath, 0755)
}

func (c *WSLSftpClient) Open(p string) (interfaces.SFTPFile, error) {
	localPath := c.mapPath(p)
	return os.Open(localPath)
}

func (c *WSLSftpClient) Create(p string) (interfaces.SFTPFile, error) {
	localPath := c.mapPath(p)
	_ = os.MkdirAll(filepath.Dir(localPath), 0755)
	return os.Create(localPath)
}

func (c *WSLSftpClient) Getwd() (string, error) {
	return "/", nil
}

func (c *WSLSftpClient) Close() error {
	return nil
}
