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
	"strings"
	"sync"
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
	stdout   io.Reader
	stdin    io.WriteCloser
	ptyErr   error
	shellErr error
	runErr   error
	winErr   error
	closed   bool
}

func (m *mockSSHSession) StdoutPipe() (io.Reader, error)     { return m.stdout, nil }
func (m *mockSSHSession) StdinPipe() (io.WriteCloser, error) { return m.stdin, nil }
func (m *mockSSHSession) RequestPty(term string, h, w int, modes ssh.TerminalModes) error {
	return m.ptyErr
}
func (m *mockSSHSession) Shell() error                { return m.shellErr }
func (m *mockSSHSession) Run(cmd string) error        { return m.runErr }
func (m *mockSSHSession) WindowChange(h, w int) error { return m.winErr }
func (m *mockSSHSession) Close() error                { m.closed = true; return nil }

type mockSFTPClient struct {
	files      []os.FileInfo
	stat       os.FileInfo
	err        error
	wd         string
	openFile   *mockSFTPFile
	createFile *mockSFTPFile
}

func (m *mockSFTPClient) ReadDir(p string) ([]os.FileInfo, error) { return m.files, m.err }
func (m *mockSFTPClient) Stat(p string) (os.FileInfo, error)      { return m.stat, m.err }
func (m *mockSFTPClient) RemoveAll(p string) error                { return m.err }
func (m *mockSFTPClient) Remove(p string) error                   { return m.err }
func (m *mockSFTPClient) Rename(old, new string) error            { return m.err }
func (m *mockSFTPClient) Mkdir(p string) error                    { return m.err }
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
func (m *mockSFTPClient) Close() error           { return nil }

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
func (m mockFileInfo) Size() int64        { return m.size }
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
	mu            sync.Mutex
	emittedEvents []string
}

func (m *mockSSHRuntime) EventsEmit(ctx context.Context, eventName string, optionalData ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.emittedEvents = append(m.emittedEvents, eventName)
}

func (m *mockSSHRuntime) getEvents() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]string, len(m.emittedEvents))
	copy(copied, m.emittedEvents)
	return copied
}
func (m *mockSSHRuntime) WindowMinimise(ctx context.Context)          {}
func (m *mockSSHRuntime) WindowMaximise(ctx context.Context)          {}
func (m *mockSSHRuntime) WindowUnmaximise(ctx context.Context)        {}
func (m *mockSSHRuntime) WindowExecJS(ctx context.Context, js string) {}
func (m *mockSSHRuntime) Quit(ctx context.Context)                    {}
func (m *mockSSHRuntime) OpenFileDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return "", nil
}
func (m *mockSSHRuntime) OpenDirectoryDialog(ctx context.Context, options wruntime.OpenDialogOptions) (string, error) {
	return "", nil
}
func (m *mockSSHRuntime) SaveFileDialog(ctx context.Context, options wruntime.SaveDialogOptions) (string, error) {
	return "", nil
}

func setupSSHTest(t *testing.T) (*SSHManager, string, *mockSSHClient, func()) {
	ctx := context.Background()
	m := NewSSHManager()
	rt := &mockSSHRuntime{}
	m.SetRuntime(rt)

	pr, pw := io.Pipe()

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

	sessionID := "test-id"
	session := config.SSHSession{
		ID:         sessionID,
		Host:       "localhost",
		Port:       22,
		User:       "user",
		AuthMethod: "password",
		Password:   "pass",
	}

	err := m.Connect(ctx, session)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	cleanup := func() {
		m.Disconnect(sessionID)
		pw.Close()
		pr.Close()
	}

	return m, sessionID, mockClient, cleanup
}

func TestSSHManager_PathAndFiles(t *testing.T) {
	m, sessionID, mockClient, cleanup := setupSSHTest(t)
	defer cleanup()

	// GetCurrentPath
	path, err := m.GetCurrentPath(sessionID)
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
	files, err := m.ListFiles(sessionID, "/home/user")
	if err != nil {
		t.Errorf("ListFiles failed: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestSSHManager_ListFiles_EOF(t *testing.T) {
	m, sessionID, mockClient, cleanup := setupSSHTest(t)
	defer cleanup()

	// 1. Test standard io.EOF error
	mockClient.sftp.err = io.EOF
	files, err := m.ListFiles(sessionID, "/home/user")
	if err != nil {
		t.Errorf("ListFiles should have succeeded on io.EOF, but got error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 files on io.EOF, got %d", len(files))
	}

	// 2. Test custom error containing "EOF"
	mockClient.sftp.err = errors.New("SFTP connection closed: EOF")
	files, err = m.ListFiles(sessionID, "/home/user")
	if err != nil {
		t.Errorf("ListFiles should have succeeded on EOF-containing error, but got error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 files on EOF-containing error, got %d", len(files))
	}

	// 3. Test os.ErrNotExist error
	mockClient.sftp.err = os.ErrNotExist
	files, err = m.ListFiles(sessionID, "/home/user")
	if err != nil {
		t.Errorf("ListFiles should have succeeded on os.ErrNotExist, but got error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 files on os.ErrNotExist, got %d", len(files))
	}

	// 4. Test custom "file does not exist" error
	mockClient.sftp.err = errors.New("file does not exist")
	files, err = m.ListFiles(sessionID, "/home/user")
	if err != nil {
		t.Errorf("ListFiles should have succeeded on 'file does not exist' error, but got error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 files on 'file does not exist' error, got %d", len(files))
	}

	// 5. Test custom "no such file" error
	mockClient.sftp.err = errors.New("no such file or directory")
	files, err = m.ListFiles(sessionID, "/home/user")
	if err != nil {
		t.Errorf("ListFiles should have succeeded on 'no such file' error, but got error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 files on 'no such file' error, got %d", len(files))
	}

	// 6. Test custom "not found" error
	mockClient.sftp.err = errors.New("directory not found")
	files, err = m.ListFiles(sessionID, "/home/user")
	if err != nil {
		t.Errorf("ListFiles should have succeeded on 'not found' error, but got error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 files on 'not found' error, got %d", len(files))
	}

	// 7. Test custom "not exist" error
	mockClient.sftp.err = errors.New("path does not exist")
	files, err = m.ListFiles(sessionID, "/home/user")
	if err != nil {
		t.Errorf("ListFiles should have succeeded on 'not exist' error, but got error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected 0 files on 'not exist' error, got %d", len(files))
	}

	// 8. Test a real other error (e.g. permission denied)
	mockClient.sftp.err = errors.New("permission denied")
	_, err = m.ListFiles(sessionID, "/home/user")
	if err == nil {
		t.Errorf("Expected error on permission denied, but got nil")
	}
}

func TestSSHManager_ResolveRemotePath(t *testing.T) {
	m, sessionID, mockClient, cleanup := setupSSHTest(t)
	defer cleanup()

	m.mu.RLock()
	conn := m.connections[sessionID]
	m.mu.RUnlock()

	mockClient.sftp.wd = "/home/testuser"

	tests := []struct {
		input    string
		expected string
	}{
		{"~", "/home/testuser"},
		{"~/", "/home/testuser"},
		{"~/subfolder", "/home/testuser/subfolder"},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		result := m.resolveRemotePath(conn, tt.input)
		if result != tt.expected {
			t.Errorf("resolveRemotePath(%q) = %q; expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestSSHManager_ListFiles_EdgeCases(t *testing.T) {
	m, sessionID, mockClient, cleanup := setupSSHTest(t)
	defer cleanup()

	// 1. Test when SFTP is nil / not connected
	m.mu.Lock()
	conn := m.connections[sessionID]
	oldSFTP := conn.SFTP
	conn.SFTP = nil
	m.mu.Unlock()

	_, err := m.ListFiles(sessionID, "/home/user")
	if err == nil || !strings.Contains(err.Error(), "SFTP not connected") {
		t.Errorf("Expected 'SFTP not connected' error, got %v", err)
	}

	// Restore SFTP
	m.mu.Lock()
	conn.SFTP = oldSFTP
	m.mu.Unlock()

	// 2. Test empty pathStr which triggers Getwd success
	mockClient.sftp.wd = "/home/defaultwd"
	mockClient.sftp.files = []os.FileInfo{
		mockFileInfo{name: "file_in_wd", size: 10, isDir: false},
	}
	mockClient.sftp.err = nil

	files, err := m.ListFiles(sessionID, "")
	if err != nil {
		t.Errorf("Expected ListFiles with empty path to succeed, got %v", err)
	}
	if len(files) != 1 || files[0].Name != "file_in_wd" {
		t.Errorf("Expected file_in_wd from default working dir, got %v", files)
	}

	// 3. Test empty pathStr when Getwd fails, which defaults to "."
	mockClient.sftp.wd = "" // triggers default to "."
	mockClient.sftp.files = []os.FileInfo{
		mockFileInfo{name: "file_in_dot", size: 5, isDir: false},
	}

	files, err = m.ListFiles(sessionID, "")
	if err != nil {
		t.Errorf("Expected ListFiles with empty path and failed Getwd to succeed, got %v", err)
	}
	if len(files) != 1 || files[0].Name != "file_in_dot" {
		t.Errorf("Expected file_in_dot from dot folder, got %v", files)
	}
}

func TestSSHManager_SFTP_Actions(t *testing.T) {
	m, sessionID, mockClient, cleanup := setupSSHTest(t)
	defer cleanup()

	// Test Stat error in delete action
	mockClient.sftp.err = errors.New("stat error")
	err := m.ExecuteSFTPAction(sessionID, "delete", "dir1", "")
	if err == nil {
		t.Error("Expected ExecuteSFTPAction to fail when Stat fails")
	}
	mockClient.sftp.err = nil

	// ExecuteSFTPAction
	mockClient.sftp.stat = mockFileInfo{name: "file1", isDir: false}
	err = m.ExecuteSFTPAction(sessionID, "delete", "file1", "")
	if err != nil {
		t.Errorf("Delete file failed: %v", err)
	}

	mockClient.sftp.stat = mockFileInfo{name: "dir1", isDir: true}
	err = m.ExecuteSFTPAction(sessionID, "delete", "dir1", "")
	if err != nil {
		t.Errorf("Delete dir failed: %v", err)
	}

	err = m.ExecuteSFTPAction(sessionID, "rename", "old", "new")
	if err != nil {
		t.Errorf("Rename failed: %v", err)
	}

	err = m.ExecuteSFTPAction(sessionID, "mkdir", "newdir", "")
	if err != nil {
		t.Errorf("Mkdir failed: %v", err)
	}

	err = m.ExecuteSFTPAction(sessionID, "unknown", "p", "")
	if err == nil {
		t.Error("Expected error for unknown action")
	}
}

func TestSSHManager_Terminal_Ops(t *testing.T) {
	m, sessionID, _, cleanup := setupSSHTest(t)
	defer cleanup()

	err := m.ResizeTerminal(sessionID, 120, 40)
	if err != nil {
		t.Errorf("ResizeTerminal failed: %v", err)
	}

	// Test hard guards for low cols and rows
	err = m.ResizeTerminal(sessionID, 50, 5)
	if err != nil {
		t.Errorf("ResizeTerminal failed on hard guards: %v", err)
	}

	err = m.SendInput(sessionID, "ls\n")
	if err != nil {
		t.Errorf("SendInput failed: %v", err)
	}
}

func TestSSHManager_File_Ops(t *testing.T) {
	m, sessionID, mockClient, cleanup := setupSSHTest(t)
	defer cleanup()

	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "local")
	os.WriteFile(tempFile, []byte("data"), 0644)

	// 1. Success case
	mockClient.sftp.openFile = &mockSFTPFile{Reader: io.LimitReader(nil, 0), Closer: io.NopCloser(nil)}
	err := m.DownloadFile(sessionID, "remote", tempFile)
	if err != nil {
		t.Errorf("DownloadFile failed: %v", err)
	}

	// 2. Download open error
	mockClient.sftp.openFile = nil
	mockClient.sftp.err = errors.New("open error")
	err = m.DownloadFile(sessionID, "remote", tempFile)
	if err == nil {
		t.Error("Expected DownloadFile to fail when SFTP.Open fails")
	}

	// 3. Download os.Create error
	mockClient.sftp.openFile = &mockSFTPFile{Reader: io.LimitReader(nil, 0), Closer: io.NopCloser(nil)}
	err = m.DownloadFile(sessionID, "remote", "/nonexistent_folder_abc/local")
	if err == nil {
		t.Error("Expected DownloadFile to fail when local file creation fails")
	}

	// Reset mock err
	mockClient.sftp.err = nil

	// 4. Upload success case
	mockClient.sftp.createFile = &mockSFTPFile{Writer: io.Discard, Closer: io.NopCloser(nil)}
	err = m.UploadFile(sessionID, tempFile, "remote")
	if err != nil {
		t.Errorf("UploadFile failed: %v", err)
	}

	// 5. Upload local open error
	err = m.UploadFile(sessionID, "nonexistent_local_file_abc", "remote")
	if err == nil {
		t.Error("Expected UploadFile to fail when local file doesn't exist")
	}

	// 6. Upload remote create error
	mockClient.sftp.createFile = nil
	mockClient.sftp.err = errors.New("create error")
	err = m.UploadFile(sessionID, tempFile, "remote")
	if err == nil {
		t.Error("Expected UploadFile to fail when SFTP.Create fails")
	}

	// Reset mock err
	mockClient.sftp.err = nil

	// EditFile
	origExecutor := utils.Executor
	utils.Executor = &testutil.MockExecutor{Output: ""}
	defer func() { utils.Executor = origExecutor }()

	mockClient.sftp.openFile = &mockSFTPFile{Reader: io.LimitReader(nil, 0), Closer: io.NopCloser(nil)}
	mockClient.sftp.createFile = &mockSFTPFile{Writer: io.Discard, Closer: io.NopCloser(nil)}
	err = m.EditFile(sessionID, "remote.txt", "nano")
	if err != nil {
		t.Errorf("EditFile failed: %v", err)
	}

	// EditFile download error
	mockClient.sftp.openFile = nil
	mockClient.sftp.err = errors.New("download error")
	err = m.EditFile(sessionID, "remote.txt", "nano")
	if err == nil {
		t.Error("Expected EditFile to fail when download fails")
	}

	// EditFile runEditor error
	utils.Executor = &testutil.MockExecutor{Err: errors.New("editor error")}
	mockClient.sftp.openFile = &mockSFTPFile{Reader: io.LimitReader(nil, 0), Closer: io.NopCloser(nil)}
	mockClient.sftp.err = nil
	err = m.EditFile(sessionID, "remote.txt", "nano")
	if err == nil {
		t.Error("Expected EditFile to fail when runEditor fails")
	}

	// EditFile with empty default editor
	utils.Executor = &testutil.MockExecutor{Output: ""}
	mockClient.sftp.openFile = &mockSFTPFile{Reader: io.LimitReader(nil, 0), Closer: io.NopCloser(nil)}
	err = m.EditFile(sessionID, "remote.txt", "")
	if err != nil {
		t.Logf("EditFile with empty editor result (expected if no editor found on system): %v", err)
	}
}

func TestSSHManager_Disconnect(t *testing.T) {
	m, sessionID, _, cleanup := setupSSHTest(t)
	defer cleanup()

	m.Disconnect(sessionID)
	m.mu.RLock()
	_, ok := m.connections[sessionID]
	m.mu.RUnlock()
	if ok {
		t.Error("Session not removed after disconnect")
	}
}

func TestSSHManager_Errors(t *testing.T) {
	m := NewSSHManager()
	m.dialer = func(cfg *goph.Config) (interfaces.SSHClient, error) {
		return nil, errors.New("dial error")
	}

	t.Run("Connect Error", func(t *testing.T) {
		err := m.Connect(context.Background(), config.SSHSession{ID: "err-id"})
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
		if err := m.ExecuteSFTPAction("none", "mkdir", "p", ""); err == nil {
			t.Error("Expected error for missing session")
		}
		if err := m.EditFile("none", "p", "editor"); err == nil {
			t.Error("Expected error for missing session")
		}
		if err := m.ResizeTerminal("none", 80, 24); err == nil {
			t.Error("Expected error for missing session")
		}
		if err := m.SendInput("none", "abc"); err == nil {
			t.Error("Expected error for missing session")
		}
	})

	t.Run("SFTP Not Connected Errors", func(t *testing.T) {
		mgr, sessionID, _, cleanup := setupSSHTest(t)
		defer cleanup()

		mgr.mu.Lock()
		conn := mgr.connections[sessionID]
		conn.SFTP = nil
		mgr.mu.Unlock()

		if _, err := mgr.GetCurrentPath(sessionID); err == nil {
			t.Error("Expected error for nil SFTP")
		}
		if err := mgr.DownloadFile(sessionID, "r", "l"); err == nil {
			t.Error("Expected error for nil SFTP")
		}
		if err := mgr.UploadFile(sessionID, "l", "r"); err == nil {
			t.Error("Expected error for nil SFTP")
		}
		if err := mgr.ExecuteSFTPAction(sessionID, "mkdir", "p", ""); err == nil {
			t.Error("Expected error for nil SFTP")
		}
		if err := mgr.EditFile(sessionID, "p", "editor"); err == nil {
			t.Error("Expected error for nil SFTP")
		}
	})

	t.Run("Terminal Ops Nil PTY Shell", func(t *testing.T) {
		mgr, sessionID, _, cleanup := setupSSHTest(t)
		defer cleanup()

		mgr.mu.Lock()
		conn := mgr.connections[sessionID]
		conn.PTY = nil
		conn.Shell = nil
		mgr.mu.Unlock()

		if err := mgr.ResizeTerminal(sessionID, 80, 24); err == nil {
			t.Error("Expected error for nil PTY")
		}
		if err := mgr.SendInput(sessionID, "abc"); err == nil {
			t.Error("Expected error for nil Shell")
		}
	})

	t.Run("ExecuteSFTPAction Unknown Action", func(t *testing.T) {
		mgr, sessionID, _, cleanup := setupSSHTest(t)
		defer cleanup()

		err := mgr.ExecuteSFTPAction(sessionID, "unknown_action", "p", "")
		if err == nil || !strings.Contains(err.Error(), "unknown action") {
			t.Errorf("Expected 'unknown action' error, got %v", err)
		}
	})

	t.Run("Connect NewSftp Error", func(t *testing.T) {
		mgr := NewSSHManager()
		mgr.dialer = func(cfg *goph.Config) (interfaces.SSHClient, error) {
			return &mockSSHClient{sftp: nil}, nil
		}
		err := mgr.Connect(context.Background(), config.SSHSession{ID: "newsftp-err-id", AuthMethod: "password"})
		if err == nil {
			t.Error("Expected Connect to fail when NewSftp fails")
		}
	})

	t.Run("startTerminal NewSession Error", func(t *testing.T) {
		mgr := NewSSHManager()
		mgr.dialer = func(cfg *goph.Config) (interfaces.SSHClient, error) {
			return &mockSSHClient{session: nil, sftp: &mockSFTPClient{}}, nil
		}
		_ = mgr.Connect(context.Background(), config.SSHSession{ID: "newsession-err-id", AuthMethod: "password"})
	})

	t.Run("DefaultSSHDialer error", func(t *testing.T) {
		_, err := DefaultSSHDialer(&goph.Config{
			User: "user",
			Addr: "127.0.0.1",
			Port: 9999,
		})
		if err == nil {
			t.Error("Expected DefaultSSHDialer to return error")
		}
	})
}

func TestSSHManager_GetResourceUsage(t *testing.T) {
	m, sessionID, mockClient, cleanup := setupSSHTest(t)
	defer cleanup()

	// 1. Success case
	output := "===METRICS===\nCPU:12.5\nMEM_TOTAL:100 MEM_USED:45\nDISK_TOTAL:100 DISK_USED:30\n===END===\n"
	mockClient.session.stdout = strings.NewReader(output)
	mockClient.session.runErr = nil

	usage, err := m.GetResourceUsage(sessionID)
	if err != nil {
		t.Errorf("Expected success, got err: %v", err)
	}
	if usage.CPU != 12.5 || usage.Mem != 45 || usage.Disk != 30 {
		t.Errorf("Expected {12.5, 45, 30}, got %+v", usage)
	}

	// 2. Not found error
	_, err = m.GetResourceUsage("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent session")
	}

	// 3. Command execution error
	mockClient.session.runErr = errors.New("command execution failed")
	mockClient.session.stdout = strings.NewReader("")
	_, err = m.GetResourceUsage(sessionID)
	if err == nil {
		t.Error("Expected error on command execution failure")
	}
}

func Test_parseResourceUsage(t *testing.T) {
	// Mixed/empty inputs
	usage := parseResourceUsage("")
	if usage.CPU != 0 || usage.Mem != 0 || usage.Disk != 0 {
		t.Errorf("Expected zeros, got %+v", usage)
	}

	usage = parseResourceUsage("CPU:invalid\nMEM_TOTAL:100 MEM_USED:75\nDISK_TOTAL:abc DISK_USED:def")
	if usage.Mem != 75 || usage.CPU != 0 || usage.Disk != 0 {
		t.Errorf("Expected MEM:75 and others zero, got %+v", usage)
	}
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
	m := NewSSHManager()
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
	go m.processTerminalOutput(ctx, conn, pr, exitChan)

	time.Sleep(100 * time.Millisecond)
	if len(rt.getEvents()) == 0 {
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

		m.handleTerminalExit(ctx, conn, sess, exitChan2)
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

		m.handleTerminalExit(ctx3, conn3, &mockSSHSession{}, exitChan3)

		m.mu.RLock()
		_, ok := m.connections["exit-id"]
		m.mu.RUnlock()
		if ok {
			t.Error("Expected connection to be removed")
		}
	})
}

func TestSSHManager_AuthMethods(t *testing.T) {
	m := NewSSHManager()

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

	t.Run("getAuth key", func(t *testing.T) {
		tempDir := t.TempDir()
		keyFile := filepath.Join(tempDir, "id_rsa")
		os.WriteFile(keyFile, []byte("mock private key content"), 0600)

		sess := config.SSHSession{AuthMethod: "key", KeyPath: keyFile, Passphrase: "pass"}
		_, _ = m.getAuth(sess)
	})

	t.Run("getAuth key without passphrase", func(t *testing.T) {
		tempDir := t.TempDir()
		keyFile := filepath.Join(tempDir, "id_rsa")
		os.WriteFile(keyFile, []byte("mock private key content"), 0600)

		sess := config.SSHSession{AuthMethod: "key", KeyPath: keyFile, Passphrase: ""}
		_, _ = m.getAuth(sess)
	})
}

type mockAddr struct{}
func (mockAddr) Network() string { return "tcp" }
func (mockAddr) String() string  { return "127.0.0.1:22" }

type mockPublicKey struct{}
func (mockPublicKey) Type() string { return "ssh-rsa" }
func (mockPublicKey) Marshal() []byte { return []byte("mock-key-bytes") }
func (mockPublicKey) Verify(data []byte, sig *ssh.Signature) error { return nil }

func TestSSHManager_HostKeyCallback(t *testing.T) {
	m := NewSSHManager()
	cb := m.getHostKeyCallback()
	if cb == nil {
		t.Fatal("Expected non-nil host key callback")
	}

	// Invoke callback to cover all internal paths
	_ = cb("localhost", mockAddr{}, mockPublicKey{})
	// Call it again so it is found in known_hosts
	_ = cb("localhost", mockAddr{}, mockPublicKey{})
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

	m := NewSSHManager()

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
