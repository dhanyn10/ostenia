package service

import (
	"context"
	"io"
	"os"
	"os/exec"
	"ostenia/internal/config"
	"path/filepath"
	"strings"
	"testing"
)

func TestWSLClient_SessionAndShell(t *testing.T) {
	// Set mock RuntimeGOOS to test both flows
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "linux"

	// Mock wslCommand to run echo/sh instead of hanging
	origWslCommand := wslCommand
	defer func() { wslCommand = origWslCommand }()
	wslCommand = func(distro string, args ...string) *exec.Cmd {
		// Mock a command that echos inputs or outputs custom text
		script := "echo 'wsl-mock-session'"
		if len(args) > 0 {
			script = "echo 'wsl-mock-args: " + strings.Join(args, " ") + "'"
		}
		return exec.Command("sh", "-c", script)
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

	// Request Pty should be a no-op
	if err := sess.RequestPty("xterm", 24, 80, nil); err != nil {
		t.Errorf("RequestPty failed: %v", err)
	}

	// WindowChange should be a no-op
	if err := sess.WindowChange(24, 80); err != nil {
		t.Errorf("WindowChange failed: %v", err)
	}

	// Shell should start the mock command
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

func TestWSLClient_SessionWithCustomUser(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "linux"

	origWslCommand := wslCommand
	defer func() { wslCommand = origWslCommand }()

	var commandArgs []string
	wslCommand = func(distro string, args ...string) *exec.Cmd {
		commandArgs = args
		return exec.Command("sh", "-c", "echo 'ok'")
	}

	client := NewWSLClient("Ubuntu", "customuser")
	sess, err := client.NewSession()
	if err != nil {
		t.Fatalf("Failed to create WSL session: %v", err)
	}

	if err := sess.Shell(); err != nil {
		t.Fatalf("Shell failed: %v", err)
	}

	// The WSL Session should pass "-u" and "customuser" to the command
	if len(commandArgs) < 2 || commandArgs[0] != "-u" || commandArgs[1] != "customuser" {
		t.Errorf("Expected user arguments in wslCommand, got %v", commandArgs)
	}
	sess.Close()
}

func TestWSLClient_SftpAndFiles(t *testing.T) {
	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()
	RuntimeGOOS = "linux"

	tempDir := t.TempDir()
	origWslRootPath := wslRootPath
	defer func() { wslRootPath = origWslRootPath }()
	wslRootPath = func(distro string) string {
		return filepath.Join(tempDir, distro)
	}

	client := NewWSLClient("Ubuntu-Test", "")
	sftpClient, err := client.NewSftp()
	if err != nil {
		t.Fatalf("Failed to create WSL SFTP Client: %v", err)
	}
	defer sftpClient.Close()

	// 1. Getwd
	wd, err := sftpClient.Getwd()
	if err != nil || wd != "/" {
		t.Errorf("Expected Getwd to return '/', got '%s' (err: %v)", wd, err)
	}

	// 2. Mkdir
	err = sftpClient.Mkdir("/testdir")
	if err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	// Verify local dir was created
	localDir := filepath.Join(tempDir, "Ubuntu-Test", "testdir")
	if info, err := os.Stat(localDir); err != nil || !info.IsDir() {
		t.Errorf("Expected local directory to be created at %s, but err: %v", localDir, err)
	}

	// 3. Create file
	file, err := sftpClient.Create("/testdir/file.txt")
	if err != nil {
		t.Fatalf("Create file failed: %v", err)
	}
	_, err = file.Write([]byte("hello wsl sftp"))
	if err != nil {
		t.Fatalf("Write file failed: %v", err)
	}
	file.Close()

	// 4. Open and Read file
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

	// 5. Stat
	info, err := sftpClient.Stat("/testdir/file.txt")
	if err != nil || info.IsDir() {
		t.Errorf("Stat failed or is directory: %v", err)
	}

	// 6. ReadDir
	infos, err := sftpClient.ReadDir("/testdir")
	if err != nil || len(infos) != 1 || infos[0].Name() != "file.txt" {
		t.Errorf("ReadDir failed or length mismatch: %v (infos: %v)", err, infos)
	}

	// 7. Rename
	err = sftpClient.Rename("/testdir/file.txt", "/testdir/file_new.txt")
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}
	// Verify rename
	if _, err := os.Stat(filepath.Join(tempDir, "Ubuntu-Test", "testdir", "file_new.txt")); err != nil {
		t.Errorf("Expected renamed file, got stat err: %v", err)
	}

	// 8. Remove
	err = sftpClient.Remove("/testdir/file_new.txt")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// 9. RemoveAll
	err = sftpClient.RemoveAll("/testdir")
	if err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	// Verify cleanup
	if _, err := os.Stat(localDir); !os.IsNotExist(err) {
		t.Errorf("Expected directory to be removed, but stat err was: %v", err)
	}

	client.Close()
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
			// Output "Ubuntu-22.04\nDebian\n" in UTF-16LE
			// UTF-16LE BOM: FF FE
			// 'U': 55 00, 'b': 62 00, 'u': 75 00, 'n': 6E 00, 't': 74 00, 'u': 75 00, '-': 2D 00, '2': 32 00, '2': 32 00, '.': 2E 00, '0': 30 00, '4': 34 00, '\r': 0D 00, '\n': 0A 00
			// 'D': 44 00, 'e': 65 00, 'b': 62 00, 'i': 69 00, 'a': 61 00, 'n': 6E 00, '\r': 0D 00, '\n': 0A 00
			utf16Output := []byte{
				0xFF, 0xFE, // BOM
				0x55, 0x00, 0x62, 0x00, 0x75, 0x00, 0x6E, 0x00, 0x74, 0x00, 0x75, 0x00, 0x2D, 0x00, 0x32, 0x00, 0x32, 0x00, 0x2E, 0x00, 0x30, 0x00, 0x34, 0x00, 0x0D, 0x00, 0x0A, 0x00,
				0x44, 0x00, 0x65, 0x00, 0x62, 0x00, 0x69, 0x00, 0x61, 0x00, 0x6E, 0x00, 0x0D, 0x00, 0x0A, 0x00,
			}
			// Let's write a mock sh script to output these bytes
			// We can write them to a temp file and cat them, or echo -ne hex.
			// But since we can write to a temp file:
			tempFile := filepath.Join(t.TempDir(), "utf16_output")
			_ = os.WriteFile(tempFile, utf16Output, 0644)
			return exec.Command("cat", tempFile)
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
		return exec.Command("sh", "-c", "echo 'connected'")
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

	// Verify connection exists in manager
	m.mu.RLock()
	conn, ok := m.connections["wsl-session-id"]
	m.mu.RUnlock()
	if !ok || conn == nil {
		t.Fatal("Expected connected WSL session in SSHManager connections map")
	}

	// Verify standard client is WSLSFTPClient
	_, isWSLSFTP := conn.SFTP.(*WSLSFTPClient)
	if !isWSLSFTP {
		t.Error("Expected connection SFTP client to be *WSLSFTPClient")
	}
}
