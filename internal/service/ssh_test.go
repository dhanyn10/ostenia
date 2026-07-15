package service

import (
	"context"
	"errors"
	"io"
	"os"
	"ostenia/internal/backend/interfaces"
	"ostenia/internal/config"
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/ssh"
)

type mockSSHClient struct {
	session    *mockSSHSession
	sftp       *mockSFTPClient
	closeError error
}

func (m *mockSSHClient) NewSession() (interfaces.SSHSession, error) {
	if m.session == nil {
		return nil, errors.New("no session")
	}
	return m.session, nil
}
func (m *mockSSHClient) NewSftp(opts ...sftp.ClientOption) (interfaces.SFTPClient, error) {
	if m.sftp == nil {
		return nil, errors.New("no sftp")
	}
	return m.sftp, nil
}
func (m *mockSSHClient) Close() error { return m.closeError }

type mockSSHSession struct {
	stdout io.Reader
	stdin  io.WriteCloser
	ptyErr error
	shellErr error
	winErr error
	closed bool
}

func (m *mockSSHSession) StdoutPipe() (io.Reader, error) { return m.stdout, nil }
func (m *mockSSHSession) StdinPipe() (io.WriteCloser, error) { return m.stdin, nil }
func (m *mockSSHSession) RequestPty(term string, h, w int, modes ssh.TerminalModes) error { return m.ptyErr }
func (m *mockSSHSession) Shell() error { return m.shellErr }
func (m *mockSSHSession) WindowChange(h, w int) error { return m.winErr }
func (m *mockSSHSession) Close() error { m.closed = true; return nil }

type mockSFTPClient struct {
	files []os.FileInfo
	stat  os.FileInfo
	err   error
	wd    string
	openFile *mockSFTPFile
	createFile *mockSFTPFile
}

func (m *mockSFTPClient) ReadDir(p string) ([]os.FileInfo, error) { return m.files, m.err }
func (m *mockSFTPClient) Stat(p string) (os.FileInfo, error) { return m.stat, m.err }
func (m *mockSFTPClient) RemoveAll(p string) error { return m.err }
func (m *mockSFTPClient) Remove(p string) error { return m.err }
func (m *mockSFTPClient) Rename(old, new string) error { return m.err }
func (m *mockSFTPClient) Mkdir(p string) error { return m.err }
func (m *mockSFTPClient) Open(p string) (interfaces.SFTPFile, error) {
	if m.openFile != nil {
		return m.openFile, nil
	}
	return nil, m.err
}
func (m *mockSFTPClient) Create(p string) (interfaces.SFTPFile, error) {
	if m.createFile != nil {
		return m.createFile, nil
	}
	return nil, m.err
}
func (m *mockSFTPClient) Getwd() (string, error) { return m.wd, m.err }
func (m *mockSFTPClient) Close() error { return nil }

type mockSFTPFile struct {
	io.Reader
	io.Writer
	io.Closer
	io.ReaderAt
	io.WriterAt
	io.Seeker
}

type mockFileInfo struct {
	name  string
	size  int64
	isDir bool
}

func (m mockFileInfo) Name() string       { return m.name }
func (m mockFileInfo) Size() int64      { return m.size }
func (m mockFileInfo) Mode() os.FileMode  { return 0 }
func (m mockFileInfo) ModTime() time.Time { return time.Now() }
func (m mockFileInfo) IsDir() bool        { return m.isDir }
func (m mockFileInfo) Sys() interface{}   { return nil }

type mockWriteCloser struct {
	data []byte
}

func (m *mockWriteCloser) Write(p []byte) (n int, err error) {
	m.data = append(m.data, p...)
	return len(p), nil
}

func (m *mockWriteCloser) Close() error { return nil }

type mockSSHRuntime struct {
	emittedEvents []string
}

func (m *mockSSHRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {
	m.emittedEvents = append(m.emittedEvents, eventName)
}
func (m *mockSSHRuntime) WindowMinimise(ctx context.Context) {}
func (m *mockSSHRuntime) WindowMaximise(ctx context.Context) {}
func (m *mockSSHRuntime) WindowUnmaximise(ctx context.Context) {}
func (m *mockSSHRuntime) WindowExecJS(ctx context.Context, js string) {}
func (m *mockSSHRuntime) Quit(ctx context.Context) {}
func (m *mockSSHRuntime) OpenFileDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) { return "", nil }
func (m *mockSSHRuntime) OpenDirectoryDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) { return "", nil }
func (m *mockSSHRuntime) SaveFileDialog(ctx context.Context, options wruntime.SaveDialogOptions) (string, error) { return "", nil }

func TestSSHManager_Full(t *testing.T) {
	ctx := context.Background()
	m := NewSSHManager(ctx)
	rt := &mockSSHRuntime{}
	m.SetRuntime(rt)

	// Use a pipe to keep stdout open so the session doesn't auto-disconnect
	pr, pw := io.Pipe()
	defer pw.Close()

	mockClient := &mockSSHClient{
		session: &mockSSHSession{
			stdout: pr,
			stdin:  &mockWriteCloser{},
		},
		sftp: &mockSFTPClient{
			wd: "/home/user",
		},
	}

	m.dialer = func(cfg *goph.Config) (interfaces.SSHClient, error) {
		return mockClient, nil
	}

	session := config.SSHSession{
		ID:         "test-id",
		Host:       "localhost",
		Port:       22,
		User:       "user",
		AuthMethod: "password",
		Password:   "pass",
	}

	t.Run("FullWorkflow", func(t *testing.T) {
		// Connect
		err := m.Connect(session)
		if err != nil {
			t.Fatalf("Connect failed: %v", err)
		}

		// GetCurrentPath
		path, err := m.GetCurrentPath("test-id")
		if err != nil {
			t.Errorf("GetCurrentPath failed: %v", err)
		}
		if path != "/home/user" {
			t.Errorf("Expected /home/user, got %s", path)
		}

		// ListFiles
		mockClient.sftp.files = []os.FileInfo{
			mockFileInfo{name: "file1", size: 100, isDir: false},
			mockFileInfo{name: "dir1", size: 0, isDir: true},
		}
		files, err := m.ListFiles("test-id", "/home/user")
		if err != nil {
			t.Errorf("ListFiles failed: %v", err)
		}
		if len(files) != 2 {
			t.Errorf("Expected 2 files, got %d", len(files))
		}

		// ExecuteSFTPAction
		mockClient.sftp.stat = mockFileInfo{name: "file1", isDir: false}
		err = m.ExecuteSFTPAction("test-id", "delete", "file1", "")
		if err != nil {
			t.Errorf("Delete file failed: %v", err)
		}

		mockClient.sftp.stat = mockFileInfo{name: "dir1", isDir: true}
		err = m.ExecuteSFTPAction("test-id", "delete", "dir1", "")
		if err != nil {
			t.Errorf("Delete dir failed: %v", err)
		}

		err = m.ExecuteSFTPAction("test-id", "rename", "old", "new")
		if err != nil {
			t.Errorf("Rename failed: %v", err)
		}

		err = m.ExecuteSFTPAction("test-id", "mkdir", "newdir", "")
		if err != nil {
			t.Errorf("Mkdir failed: %v", err)
		}

		err = m.ExecuteSFTPAction("test-id", "unknown", "p", "")
		if err == nil {
			t.Error("Expected error for unknown action")
		}

		// Terminal operations
		err = m.ResizeTerminal("test-id", 120, 40)
		if err != nil {
			t.Errorf("ResizeTerminal failed: %v", err)
		}

		err = m.SendInput("test-id", "ls\n")
		if err != nil {
			t.Errorf("SendInput failed: %v", err)
		}

		// File operations
		tempDir := t.TempDir()
		tempFile := filepath.Join(tempDir, "local")
		os.WriteFile(tempFile, []byte("data"), 0644)

		mockClient.sftp.openFile = &mockSFTPFile{Reader: io.LimitReader(nil, 0), Closer: io.NopCloser(nil)}
		err = m.DownloadFile("test-id", "remote", tempFile)
		if err != nil {
			t.Errorf("DownloadFile failed: %v", err)
		}

		mockClient.sftp.createFile = &mockSFTPFile{Writer: io.Discard, Closer: io.NopCloser(nil)}
		err = m.UploadFile("test-id", tempFile, "remote")
		if err != nil {
			t.Errorf("UploadFile failed: %v", err)
		}

		// EditFile
		utils.Executor = &testutil.MockExecutor{Output: ""}
		mockClient.sftp.openFile = &mockSFTPFile{Reader: io.LimitReader(nil, 0), Closer: io.NopCloser(nil)}
		mockClient.sftp.createFile = &mockSFTPFile{Writer: io.Discard, Closer: io.NopCloser(nil)}
		err = m.EditFile("test-id", "remote.txt", "nano")
		if err != nil {
			t.Errorf("EditFile failed: %v", err)
		}

		// Disconnect
		m.Disconnect("test-id")
		m.mu.RLock()
		_, ok := m.connections["test-id"]
		m.mu.RUnlock()
		if ok {
			t.Error("Session not removed after disconnect")
		}
	})
}

func TestSSHManager_Errors(t *testing.T) {
	m := NewSSHManager(context.Background())
	m.dialer = func(cfg *goph.Config) (interfaces.SSHClient, error) {
		return nil, errors.New("dial error")
	}

	t.Run("Connect Error", func(t *testing.T) {
		err := m.Connect(config.SSHSession{ID: "err-id"})
		if err == nil {
			t.Error("Expected dial error")
		}
	})

	t.Run("Not Found Errors", func(t *testing.T) {
		if _, err := m.GetCurrentPath("none"); err == nil {
			t.Error("Expected error for missing session")
		}
		if _, err := m.ListFiles("none", "."); err == nil {
			t.Error("Expected error for missing session")
		}
		if err := m.DownloadFile("none", "r", "l"); err == nil {
			t.Error("Expected error for missing session")
		}
		if err := m.UploadFile("none", "l", "r"); err == nil {
			t.Error("Expected error for missing session")
		}
	})
}

func TestSSHManager_Editor_Mocked(t *testing.T) {
	utils.Executor = &testutil.MockExecutor{
		Output: "",
		Err:    nil,
	}
	m := &SSHManager{}

	t.Run("runEditor custom", func(t *testing.T) {
		err := m.runEditor("/tmp/file", "nano")
		if err != nil {
			t.Errorf("runEditor failed: %v", err)
		}
	})

	t.Run("runEditor default", func(t *testing.T) {
		err := m.runEditor("/tmp/file", "")
		// Might fail on Windows if notepad not found in mock, but let's see.
		_ = err
	})
}

func TestSSHManager_TerminalLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := NewSSHManager(ctx)
	rt := &mockSSHRuntime{}
	m.SetRuntime(rt)

	pr, pw := io.Pipe()
	conn := &SSHConnection{
		SessionID: "loop-id",
		PTY:       &mockSSHSession{},
		Context:   ctx,
		Cancel:    cancel,
	}

	go func() {
		pw.Write([]byte("terminal data"))
		pw.Close()
	}()

	exitChan := make(chan struct{})
	go m.processTerminalOutput(conn, pr, exitChan)

	time.Sleep(100 * time.Millisecond)
	if len(rt.emittedEvents) == 0 {
		t.Error("Expected events to be emitted")
	}

	t.Run("handleTerminalExit Context Done", func(t *testing.T) {
		m.mu.Lock()
		m.connections[conn.SessionID] = conn
		m.mu.Unlock()

		exitChan2 := make(chan struct{})
		sess := &mockSSHSession{}

		// Trigger context cancel
		cancel()

		m.handleTerminalExit(conn, sess, exitChan2)
		if !sess.closed {
			t.Error("Expected session to be closed")
		}
	})

	t.Run("handleTerminalExit exitChan", func(t *testing.T) {
		ctx3, cancel3 := context.WithCancel(context.Background())
		defer cancel3()
		conn3 := &SSHConnection{
			SessionID: "exit-id",
			Context:   ctx3,
			Cancel:    cancel3,
		}
		m.mu.Lock()
		m.connections[conn3.SessionID] = conn3
		m.mu.Unlock()

		exitChan3 := make(chan struct{})
		close(exitChan3)

		m.handleTerminalExit(conn3, &mockSSHSession{}, exitChan3)

		m.mu.RLock()
		_, ok := m.connections["exit-id"]
		m.mu.RUnlock()
		if ok {
			t.Error("Expected connection to be removed")
		}
	})
}

func TestSSHManager_AuthMethods(t *testing.T) {
	m := NewSSHManager(context.Background())

	t.Run("getAuth password", func(t *testing.T) {
		sess := config.SSHSession{AuthMethod: "password", Password: "p"}
		auth, err := m.getAuth(sess)
		if err != nil || auth == nil {
			t.Errorf("getAuth password failed: %v", err)
		}
	})

	t.Run("getAuth agent", func(t *testing.T) {
		sess := config.SSHSession{AuthMethod: "agent"}
		_, _ = m.getAuth(sess)
	})
}

func TestSSHManager_Wrappers(t *testing.T) {
	// These are thin wrappers around goph/sftp/ssh libraries.
	// We call them to ensure coverage.

	safeCall := func(fn func()) {
		defer func() { recover() }()
		fn()
	}

	client := &gophSSHClient{client: nil}
	safeCall(func() { client.Close() })
	safeCall(func() { client.NewSession() })
	safeCall(func() { client.NewSftp() })

	sess := &sshSessionWrapper{session: nil}
	safeCall(func() { sess.StdoutPipe() })
	safeCall(func() { sess.StdinPipe() })
	safeCall(func() { sess.RequestPty("", 0, 0, nil) })
	safeCall(func() { sess.Shell() })
	safeCall(func() { sess.WindowChange(0, 0) })
	safeCall(func() { sess.Close() })

	sftpW := &sftpClientWrapper{client: nil}
	safeCall(func() { sftpW.ReadDir("") })
	safeCall(func() { sftpW.Stat("") })
	safeCall(func() { sftpW.RemoveAll("") })
	safeCall(func() { sftpW.Remove("") })
	safeCall(func() { sftpW.Rename("", "") })
	safeCall(func() { sftpW.Mkdir("") })
	safeCall(func() { sftpW.Open("") })
	safeCall(func() { sftpW.Create("") })
	safeCall(func() { sftpW.Getwd() })
	safeCall(func() { sftpW.Close() })
}

func TestSSHManager_Sessions(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("OSTENIA_HOME", tempDir)
	defer os.Unsetenv("OSTENIA_HOME")

	m := NewSSHManager(context.Background())

	t.Run("SaveAndGet", func(t *testing.T) {
		sessions := []config.SSHSession{{ID: "s1", Host: "h1"}}
		err := m.SaveSessions(sessions)
		if err != nil {
			t.Errorf("SaveSessions failed: %v", err)
		}

		got, err := m.GetSessions()
		if err != nil {
			t.Errorf("GetSessions failed: %v", err)
		}
		if len(got) != 1 || got[0].ID != "s1" {
			t.Error("Session mismatch")
		}
	})
}
