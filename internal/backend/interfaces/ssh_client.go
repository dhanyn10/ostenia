package interfaces

import (
	"io"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SSHClient interface {
	NewSession() (SSHSession, error)
	NewSftp(...sftp.ClientOption) (SFTPClient, error)
	Close() error
}

type ResourceUsage struct {
	CPU       float64 `json:"cpu"`
	Mem       float64 `json:"mem"`
	MemTotal  float64 `json:"memTotal"`
	MemUsed   float64 `json:"memUsed"`
	Disk      float64 `json:"disk"`
	DiskTotal float64 `json:"diskTotal"`
	DiskUsed  float64 `json:"diskUsed"`
}

type SSHSession interface {
	StdoutPipe() (io.Reader, error)
	StdinPipe() (io.WriteCloser, error)
	RequestPty(term string, h, w int, modes ssh.TerminalModes) error
	Shell() error
	Run(cmd string) error
	WindowChange(h, w int) error
	Close() error
}

type SFTPClient interface {
	ReadDir(p string) ([]os.FileInfo, error)
	Stat(p string) (os.FileInfo, error)
	RemoveAll(p string) error
	Remove(p string) error
	Rename(oldpath, newpath string) error
	Mkdir(p string) error
	Open(p string) (SFTPFile, error)
	Create(p string) (SFTPFile, error)
	Getwd() (string, error)
	Close() error
}

type SFTPFile interface {
	io.Reader
	io.Writer
	io.Closer
	io.ReaderAt
	io.WriterAt
	io.Seeker
}
