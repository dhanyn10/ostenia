package service

import (
	"context"
	"io"
	"os"
	"os/exec"
	"ostenia/internal/backend/interfaces"
	"ostenia/internal/config"
	"path/filepath"
	"strings"
	"testing"
)

func TestWSLClient_SessionAndShell(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "linux"

	origWslCommand := wslCommand
	defer func() { wslCommand = origWslCommand }()
	wslCommand = func(distro string, args ...string) *exec.Cmd {
		script := "echo 'wsl-mock-session'"
		if len(args) > 0 {
			script = "echo 'wsl-mock-args: " + strings.Join(args, " ") + "'"
		}
		return exec.Command("sh", "-c", script) // NOSONAR
	}

	client := NewWSLClient("Ubuntu", "root")
	if client.Distro != "Ubuntu" || client.User != "root" {
		t.Errorf("Expected Distro Ubuntu and User root, got Distro %s, User %s", client.Distro, client.User)
	}

	sess, err := client.NewSession()
	if err != nil {
		t.Fatalf("Failed to create WSL session: %v", err)
	}

	stdout, err := sess.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe failed: %v", err)
	}

	stdin, err := sess.StdinPipe()
	if err != nil {
		t.Fatalf("StdinPipe failed: %v", err)
	}
	if stdin == nil {
		t.Error("Expected non-nil stdin pipe")
	}

	if err := sess.RequestPty("xterm", 24, 80, nil); err != nil {
		t.Errorf("RequestPty failed: %v", err)
	}

	if err := sess.WindowChange(24, 80); err != nil {
		t.Errorf("WindowChange failed: %v", err)
	}

	if err := sess.Shell(); err != nil {
		t.Fatalf("Shell failed to start: %v", err)
	}

	outputBytes, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("Failed to read stdout: %v", err)
	}

	outputStr := strings.TrimSpace(string(outputBytes))
	if outputStr != "wsl-mock-session" {
		t.Errorf("Expected 'wsl-mock-session', got '%s'", outputStr)
	}

	sess.Close()
}

func TestWSLCommand_RealImplementation(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()

	t.Run("Windows wslCommand secure environment and path", func(t *testing.T) {
		RuntimeGOOS = "windows"
		cmd := wslCommand("Ubuntu-Test-HideWindow", "ls")
		if cmd == nil {
			t.Fatal("Expected non-nil command")
		}

		termFound := false
		pathFound := false
		for _, env := range cmd.Env {
			if strings.HasPrefix(env, "TERM=") {
				termFound = true
				if env != "TERM=xterm-256color" {
					t.Errorf("Expected TERM=xterm-256color, got %s", env)
				}
			}
			if strings.HasPrefix(env, "PATH=") {
				pathFound = true
				if !strings.Contains(env, "System32") {
					t.Errorf("Expected secure Windows path, got %s", env)
				}
			}
		}

		if !termFound {
			t.Error("Expected TERM environment variable to be set")
		}
		if !pathFound {
			t.Error("Expected PATH environment variable to be set")
		}
	})

	t.Run("Linux wslCommand secure environment", func(t *testing.T) {
		RuntimeGOOS = "linux"
		cmd := wslCommand("Ubuntu-Test-HideWindow", "ls")
		if cmd == nil {
			t.Fatal("Expected non-nil command")
		}

		termFound := false
		pathFound := false
		for _, env := range cmd.Env {
			if strings.HasPrefix(env, "TERM=") {
				termFound = true
			}
			if strings.HasPrefix(env, "PATH=") {
				pathFound = true
				if env != "PATH=/usr/bin:/bin:/usr/sbin:/sbin" {
					t.Errorf("Expected secure Linux path, got %s", env)
				}
			}
		}

		if !termFound {
			t.Error("Expected TERM environment variable to be set")
		}
		if !pathFound {
			t.Error("Expected PATH environment variable to be set")
		}
	})
}

func TestWSLSession_Run(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "linux"

	origWslCommand := wslCommand
	defer func() { wslCommand = origWslCommand }()
	wslCommand = func(distro string, args ...string) *exec.Cmd {
		return exec.Command("echo", "wsl-mock-run") // NOSONAR
	}

	client := NewWSLClient("Ubuntu", "root")
	sess, err := client.NewSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	defer sess.Close()

	stdout, err := sess.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe failed: %v", err)
	}

	err = sess.Run("echo hello")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	output, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if !strings.Contains(string(output), "wsl-mock-run") {
		t.Errorf("Expected output to contain 'wsl-mock-run', got %q", string(output))
	}
}

func TestWSLClient_SessionWithCustomUser(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "linux"

	origWslCommand := wslCommand
	defer func() { wslCommand = origWslCommand }()

	var commandArgs []string
	wslCommand = func(distro string, args ...string) *exec.Cmd {
		commandArgs = args
		return exec.Command("sh", "-c", "echo 'ok'") // NOSONAR
	}

	client := NewWSLClient("Ubuntu", "customuser")
	sess, err := client.NewSession()
	if err != nil {
		t.Fatalf("Failed to create WSL session: %v", err)
	}

	if err := sess.Shell(); err != nil {
		t.Fatalf("Shell failed: %v", err)
	}

	if len(commandArgs) < 2 || commandArgs[0] != "-u" || commandArgs[1] != "customuser" {
		t.Errorf("Expected user arguments in wslCommand, got %v", commandArgs)
	}
	sess.Close()
}

func setupWSLSftpTest(t *testing.T) (*WSLClient, interfaces.SFTPClient, string, func()) {
	origGOOS := RuntimeGOOS
	RuntimeGOOS = "linux"

	tempDir := t.TempDir()
	origWslRootPath := wslRootPath
	wslRootPath = func(distro string) string {
		return filepath.Join(tempDir, distro)
	}

	client := NewWSLClient("Ubuntu-Test", "")
	sftpClient, err := client.NewSftp()
	if err != nil {
		t.Fatalf("Failed to create WSL SFTP Client: %v", err)
	}

	cleanup := func() {
		sftpClient.Close()
		client.Close()
		wslRootPath = origWslRootPath
		RuntimeGOOS = origGOOS
	}

	return client, sftpClient, tempDir, cleanup
}

func TestWSLClient_SftpAndFiles_Getwd(t *testing.T) {
	_, sftpClient, _, cleanup := setupWSLSftpTest(t)
	defer cleanup()

	wd, err := sftpClient.Getwd()
	if err != nil || wd != "/" {
		t.Errorf("Expected Getwd to return '/', got '%s' (err: %v)", wd, err)
	}
}

func TestWSLClient_toWSLPath_MountPoints(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()

	RuntimeGOOS = "windows"
	p := toWSLPath(`\\wsl.localhost\Ubuntu`, "/mnt/d/koding/ostenia")
	if !strings.HasPrefix(p, "d:\\") && !strings.HasPrefix(p, "D:\\") {
		t.Errorf("Expected path mapped to D:\\, got '%s'", p)
	}

	p2 := toWSLPath(`\\wsl.localhost\Ubuntu`, "/mnt/c")
	if !strings.HasPrefix(p2, "c:\\") && !strings.HasPrefix(p2, "C:\\") {
		t.Errorf("Expected path mapped to C:\\, got '%s'", p2)
	}

	RuntimeGOOS = "linux"
	p3 := toWSLPath(`/tmp/wsl/Ubuntu`, "/mnt/d/koding/ostenia")
	if strings.Contains(p3, "d:\\") || strings.Contains(p3, "D:\\") {
		t.Errorf("Expected Linux to not map drive letter, got '%s'", p3)
	}
}

func TestWSLClient_SftpAndFiles_MkdirAndStat(t *testing.T) {
	_, sftpClient, tempDir, cleanup := setupWSLSftpTest(t)
	defer cleanup()

	err := sftpClient.Mkdir("/testdir")
	if err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	localDir := filepath.Join(tempDir, "Ubuntu-Test", "testdir")
	if info, err := os.Stat(localDir); err != nil || !info.IsDir() {
		t.Errorf("Expected local directory to be created at %s, but err: %v", localDir, err)
	}
}

func TestWSLClient_SftpAndFiles_CreateAndWrite(t *testing.T) {
	_, sftpClient, _, cleanup := setupWSLSftpTest(t)
	defer cleanup()

	err := sftpClient.Mkdir("/testdir")
	if err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	file, err := sftpClient.Create("/testdir/file.txt")
	if err != nil {
		t.Fatalf("Create file failed: %v", err)
	}
	_, err = file.Write([]byte("hello wsl sftp"))
	if err != nil {
		t.Fatalf("Write file failed: %v", err)
	}
	file.Close()

	file, err = sftpClient.Open("/testdir/file.txt")
	if err != nil {
		t.Fatalf("Open file failed: %v", err)
	}
	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Read file failed: %v", err)
	}
	file.Close()
	if string(content) != "hello wsl sftp" {
		t.Errorf("Expected 'hello wsl sftp', got '%s'", string(content))
	}
}

func TestWSLClient_SftpAndFiles_StatAndReadDir(t *testing.T) {
	_, sftpClient, _, cleanup := setupWSLSftpTest(t)
	defer cleanup()

	err := sftpClient.Mkdir("/testdir")
	if err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	file, err := sftpClient.Create("/testdir/file.txt")
	if err != nil {
		t.Fatalf("Create file failed: %v", err)
	}
	file.Close()

	info, err := sftpClient.Stat("/testdir/file.txt")
	if err != nil || info.IsDir() {
		t.Errorf("Stat failed or is directory: %v", err)
	}

	infos, err := sftpClient.ReadDir("/testdir")
	if err != nil || len(infos) != 1 || infos[0].Name() != "file.txt" {
		t.Errorf("ReadDir failed or length mismatch: %v (infos: %v)", err, infos)
	}
}

func TestWSLClient_SftpAndFiles_RenameAndRemove(t *testing.T) {
	_, sftpClient, tempDir, cleanup := setupWSLSftpTest(t)
	defer cleanup()

	err := sftpClient.Mkdir("/testdir")
	if err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	file, err := sftpClient.Create("/testdir/file.txt")
	if err != nil {
		t.Fatalf("Create file failed: %v", err)
	}
	file.Close()

	err = sftpClient.Rename("/testdir/file.txt", "/testdir/file_new.txt")
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "Ubuntu-Test", "testdir", "file_new.txt")); err != nil {
		t.Errorf("Expected renamed file, got stat err: %v", err)
	}

	err = sftpClient.Remove("/testdir/file_new.txt")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	err = sftpClient.RemoveAll("/testdir")
	if err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	localDir := filepath.Join(tempDir, "Ubuntu-Test", "testdir")
	if _, err := os.Stat(localDir); !os.IsNotExist(err) {
		t.Errorf("Expected directory to be removed, but stat err was: %v", err)
	}
}

func TestSSHManager_GetWSLDistros(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()

	m := NewSSHManager()

	t.Run("Non-Windows platform returns empty", func(t *testing.T) {
		RuntimeGOOS = "linux"
		distros, err := m.GetWSLDistros()
		if err != nil {
			t.Fatalf("GetWSLDistros failed on linux: %v", err)
		}
		if len(distros) != 0 {
			t.Errorf("Expected 0 distros on non-Windows, got %v", distros)
		}
	})

	t.Run("Windows platform decodes UTF-16LE successfully", func(t *testing.T) {
		RuntimeGOOS = "windows"

		origWslCommand := wslCommand
		defer func() { wslCommand = origWslCommand }()

		wslCommand = func(distro string, args ...string) *exec.Cmd {
			utf16Output := []byte{
				0xFF, 0xFE,
				0x55, 0x00, 0x62, 0x00, 0x75, 0x00, 0x6E, 0x00, 0x74, 0x00, 0x75, 0x00, 0x2D, 0x00, 0x32, 0x00, 0x32, 0x00, 0x2E, 0x00, 0x30, 0x00, 0x34, 0x00, 0x0D, 0x00, 0x0A, 0x00,
				0x44, 0x00, 0x65, 0x00, 0x62, 0x00, 0x69, 0x00, 0x61, 0x00, 0x6E, 0x00, 0x0D, 0x00, 0x0A, 0x00,
			}
			tempFile := filepath.Join(t.TempDir(), "utf16_output")
			_ = os.WriteFile(tempFile, utf16Output, 0644)
			return exec.Command("cat", tempFile) // NOSONAR
		}

		distros, err := m.GetWSLDistros()
		if err != nil {
			t.Fatalf("GetWSLDistros failed on windows: %v", err)
		}

		if len(distros) != 2 || distros[0] != "Ubuntu-22.04" || distros[1] != "Debian" {
			t.Errorf("Expected ['Ubuntu-22.04', 'Debian'], got %v", distros)
		}
	})
}

func TestSSHManager_WSLConnect(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "linux"

	origWslCommand := wslCommand
	defer func() { wslCommand = origWslCommand }()
	wslCommand = func(distro string, args ...string) *exec.Cmd {
		for _, arg := range args {
			if strings.Contains(arg, "readlink") {
				return exec.Command("echo", "/mnt/d/koding/ostenia") // NOSONAR
			}
		}
		return exec.Command("sh", "-c", "echo 'connected'") // NOSONAR
	}

	m := NewSSHManager()
	sess := config.SSHSession{
		ID:   "wsl-session-id",
		Name: "My WSL",
		Host: "wsl://Ubuntu",
		Port: 22,
		User: "root",
	}

	ctx := context.Background()
	err := m.Connect(ctx, sess)
	if err != nil {
		t.Fatalf("Failed to Connect to WSL session: %v", err)
	}
	defer m.Disconnect("wsl-session-id")

	m.mu.RLock()
	conn, ok := m.connections["wsl-session-id"]
	m.mu.RUnlock()
	if !ok || conn == nil {
		t.Fatal("Expected connected WSL session in SSHManager connections map")
	}

	_, isWSLSFTP := conn.SFTP.(*WSLSFTPClient)
	if !isWSLSFTP {
		t.Error("Expected connection SFTP client to be *WSLSFTPClient")
	}

	if !conn.IsWSL {
		t.Error("Expected IsWSL to be true on WSL connection")
	}

	// Test GetCurrentPath for WSL CWD query
	cwd, err := m.GetCurrentPath("wsl-session-id")
	if err != nil {
		t.Fatalf("GetCurrentPath failed: %v", err)
	}
	if cwd != "/mnt/d/koding/ostenia" {
		t.Errorf("Expected current path to be '/mnt/d/koding/ostenia', got '%s'", cwd)
	}
}

func TestSSHManager_WSLConnect_GetCurrentPath_ErrorFallback(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "linux"

	origWslCommand := wslCommand
	defer func() { wslCommand = origWslCommand }()
	// Mock wslCommand to return an error / empty output to simulate failed CWD query
	wslCommand = func(distro string, args ...string) *exec.Cmd {
		for _, arg := range args {
			if strings.Contains(arg, "readlink") {
				return exec.Command("sh", "-c", "exit 1") // NOSONAR
			}
		}
		return exec.Command("sh", "-c", "echo 'connected'") // NOSONAR
	}

	m := NewSSHManager()
	sess := config.SSHSession{
		ID:   "wsl-session-id-err",
		Name: "My WSL Err",
		Host: "wsl://Ubuntu",
		Port: 22,
		User: "root",
	}

	ctx := context.Background()
	err := m.Connect(ctx, sess)
	if err != nil {
		t.Fatalf("Failed to Connect to WSL session: %v", err)
	}
	defer m.Disconnect("wsl-session-id-err")

	// Test GetCurrentPath which should return error and NOT fallback to "/"
	_, err = m.GetCurrentPath("wsl-session-id-err")
	if err == nil {
		t.Error("Expected GetCurrentPath to return error when CWD query fails, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "could not determine terminal CWD for WSL") {
		t.Errorf("Expected error to mention CWD for WSL, got '%v'", err)
	}
}

func TestWSLClient_SftpErrors(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "linux"

	tempDir := t.TempDir()
	origWslRootPath := wslRootPath
	defer func() { wslRootPath = origWslRootPath }()
	wslRootPath = func(distro string) string {
		return filepath.Join(tempDir, distro)
	}

	client := NewWSLClient("Ubuntu-Error", "")
	sftpClient, err := client.NewSftp()
	if err != nil {
		t.Fatalf("Failed to create WSL SFTP Client: %v", err)
	}
	defer sftpClient.Close()

	// Try reading nonexistent directory
	_, err = sftpClient.ReadDir("/nonexistent")
	if err == nil {
		t.Error("Expected ReadDir of nonexistent path to return error")
	}

	// Try stat of nonexistent file
	_, err = sftpClient.Stat("/nonexistent")
	if err == nil {
		t.Error("Expected Stat of nonexistent path to return error")
	}

	// Try opening nonexistent file
	_, err = sftpClient.Open("/nonexistent")
	if err == nil {
		t.Error("Expected Open of nonexistent path to return error")
	}

	// Try removing nonexistent file
	err = sftpClient.Remove("/nonexistent")
	if err == nil {
		t.Error("Expected Remove of nonexistent path to return error")
	}

	// Try renaming nonexistent file
	err = sftpClient.Rename("/nonexistent1", "/nonexistent2")
	if err == nil {
		t.Error("Expected Rename of nonexistent paths to return error")
	}
}

func TestWSLClient_GetWSLDistros_ErrorCase(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "windows"

	origWslCommand := wslCommand
	defer func() { wslCommand = origWslCommand }()

	// Mock wslCommand to return a failing process
	wslCommand = func(distro string, args ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "exit 1") // NOSONAR
	}

	m := NewSSHManager()
	_, err := m.GetWSLDistros()
	if err == nil {
		t.Error("Expected GetWSLDistros to fail when wsl command fails")
	}
}

func TestWSLClient_SessionPipesMock(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "linux"

	client := NewWSLClient("Ubuntu", "root")
	sess, err := client.NewSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Verify we can retrieve pipes multiple times safely (returning the cached pipes)
	stdout1, err := sess.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe failed: %v", err)
	}
	stdout2, err := sess.StdoutPipe()
	if err != nil || stdout1 != stdout2 {
		t.Error("Expected identical stdout pipe instance on subsequent calls")
	}

	stdin1, err := sess.StdinPipe()
	if err != nil {
		t.Fatalf("StdinPipe failed: %v", err)
	}
	stdin2, err := sess.StdinPipe()
	if err != nil || stdin1 != stdin2 {
		t.Error("Expected identical stdin pipe instance on subsequent calls")
	}

	sess.Close()
}
