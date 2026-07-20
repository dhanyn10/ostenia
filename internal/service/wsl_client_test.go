package service

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNewWSLClient(t *testing.T) {
	// Mock the command so it does not run interactively or hang
	origWslCommand := wslCommand
	defer func() { wslCommand = origWslCommand }()

	wslCommand = func(name string, arg ...string) *exec.Cmd {
		if runtime.GOOS == "windows" {
			return exec.Command("more")
		}
		return exec.Command("cat")
	}

	client, err := NewWSLClient("Ubuntu")
	if err != nil {
		t.Fatalf("Failed to create WSLClient: %v", err)
	}

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("Failed to create WSL session: %v", err)
	}
	defer session.Close()

	// Verify we can retrieve pipes without error
	if _, err := session.StdoutPipe(); err != nil {
		t.Errorf("StdoutPipe failed: %v", err)
	}
	if _, err := session.StdinPipe(); err != nil {
		t.Errorf("StdinPipe failed: %v", err)
	}

	// RequestPty and WindowChange should not fail
	if err := session.RequestPty("xterm", 24, 80, nil); err != nil {
		t.Errorf("RequestPty failed: %v", err)
	}
	if err := session.WindowChange(24, 80); err != nil {
		t.Errorf("WindowChange failed: %v", err)
	}

	// Start shell
	if err := session.Shell(); err != nil {
		t.Fatalf("Shell start failed: %v", err)
	}

	// Test writing
	stdin, err := session.StdinPipe()
	if err != nil {
		t.Errorf("StdinPipe failed: %v", err)
	}

	_, err = stdin.Write([]byte("echo hello\n"))
	if err != nil {
		t.Errorf("Write to stdin failed: %v", err)
	}

	if err := client.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestWSLSftpClient_Mock(t *testing.T) {
	distro := "Debian"
	sftp, err := (&WSLClient{distro: distro}).NewSftp()
	if err != nil {
		t.Fatalf("Failed to create SFTP client: %v", err)
	}
	defer sftp.Close()

	// Verify working directory
	wd, err := sftp.Getwd()
	if err != nil {
		t.Errorf("Getwd failed: %v", err)
	}
	if wd != "/" {
		t.Errorf("Expected Getwd to return /, got %q", wd)
	}

	// Dynamic path mapping testing
	wslSftp := sftp.(*WSLSftpClient)
	mapped := wslSftp.mapPath("/etc/test")
	if runtime.GOOS == "windows" {
		expected := `\\wsl.localhost\Debian\etc\test`
		if mapped != expected {
			t.Errorf("Expected mapped path to be %q, got %q", expected, mapped)
		}
	} else {
		if !strings.Contains(mapped, "Debian") || !strings.Contains(mapped, "etc") {
			t.Errorf("Expected mapped path to contain Debian/etc, got %q", mapped)
		}
	}

	// Setup clean mock file operations inside the temporary system folder
	tempWSLRoot := filepath.Join(os.TempDir(), "wsl", distro)
	_ = os.RemoveAll(tempWSLRoot)
	defer os.RemoveAll(tempWSLRoot)

	// Mkdir
	err = sftp.Mkdir("/home/user/new_folder")
	if err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	// Stat of folder
	info, err := sftp.Stat("/home/user/new_folder")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Expected /home/user/new_folder to be a directory")
	}

	// Create file
	file, err := sftp.Create("/home/user/new_folder/file.txt")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	_, err = file.Write([]byte("Hello WSL Sftp!"))
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	file.Close()

	// Open file
	openFile, err := sftp.Open("/home/user/new_folder/file.txt")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	content := make([]byte, 100)
	n, err := openFile.Read(content)
	if err != nil && err != io.EOF {
		t.Fatalf("Read failed: %v", err)
	}
	if string(content[:n]) != "Hello WSL Sftp!" {
		t.Errorf("Expected content 'Hello WSL Sftp!', got %q", string(content[:n]))
	}
	openFile.Close()

	// ReadDir
	files, err := sftp.ReadDir("/home/user/new_folder")
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("Expected 1 file in ReadDir, got %d", len(files))
	} else if files[0].Name() != "file.txt" {
		t.Errorf("Expected file.txt, got %s", files[0].Name())
	}

	// Rename file
	err = sftp.Rename("/home/user/new_folder/file.txt", "/home/user/new_folder/renamed.txt")
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	// Stat renamed file
	info, err = sftp.Stat("/home/user/new_folder/renamed.txt")
	if err != nil {
		t.Fatalf("Stat of renamed file failed: %v", err)
	}
	if info.IsDir() {
		t.Errorf("Expected renamed.txt to be a file, not a directory")
	}

	// Remove file
	err = sftp.Remove("/home/user/new_folder/renamed.txt")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// RemoveAll folder
	err = sftp.RemoveAll("/home")
	if err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	// Stat should now fail
	_, err = sftp.Stat("/home")
	if err == nil {
		t.Error("Expected Stat to fail after folder removal")
	}
}
