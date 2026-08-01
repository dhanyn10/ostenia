package service

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"ostenia/internal/backend/interfaces"
	"ostenia/internal/plugins/utils"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//go:embed scripts/get_cwd.sh
var getCWDScript string

// resolveRemotePath translates tilde-style paths (~/ atau ~) to absolute remote paths based on SFTP home directory.
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

// ListFiles returns files and directories inside the specified path.
// Gracefully absorbs typical missing or empty directory SFTP errors as successful empty listings
// to prevent disruptive front-end alert popups.
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

	// Fallback to active SFTP working directory if directory path is blank
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
		// Intercept expected connection EOF or missing directory messages and return empty slice
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

// ExecuteSFTPAction handles remote mutations: 'delete' (files/recursive directories), 'rename' (move/rename), and 'mkdir'.
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

// DownloadFile copies a file from the remote host down to the local user's storage path.
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

// UploadFile transmits a file from local disk up to the targeted remote path on the host.
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

// EditFile downloads a remote file to a safe local temporary file, triggers the preferred local text editor,
// and automatically transmits modified versions back to the remote server on save.
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

	// Create a safe unique temporary folder to prevent colliding edits
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

	// Only upload back to remote if file modtime was actually updated during the edit session
	finalInfo, err := os.Stat(localPath)
	if err == nil && finalInfo.ModTime().After(initialInfo.ModTime()) {
		return m.UploadFile(sessionID, localPath, remotePath)
	}

	return nil
}

// runEditor attempts to execute the editor program configured in user preferences,
// falling back to standard OS notepad/open defaults.
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

// getCustomEditorCmd formats command arguments to call custom editors on Windows / Linux hosts securely.
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

// getDefaultEditorCmd triggers standard system-integrated text viewers (notepad, TextEdit, etc).
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

// findLinuxEditor attempts to query general Linux text editors like xdg-open, gedit, nano.
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

// GetCurrentPath retrieves the remote path string of the current SFTP working directory or active WSL/SSH shell.
// For WSL distribution shells on Windows, it queries `/proc` inside WSL without quote mangling.
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

	// If this is a WSL connection, we can query the distro process via wslCommand to read the active shell CWD.
	if conn.IsWSL {
		var cmd *exec.Cmd
		wslCli := conn.Client.(*WSLClient)
		if wslCli.User != "" && wslCli.User != "root" {
			cmd = wslCommand(wslCli.Distro, "-u", wslCli.User, "sh", "-c", getCWDScript)
		} else {
			cmd = wslCommand(wslCli.Distro, "sh", "-c", getCWDScript)
		}

		output, err := cmd.Output()
		if err == nil {
			decoded := strings.TrimSpace(decodeMaybeUTF16(output))
			if decoded != "" && strings.HasPrefix(decoded, "/") {
				return decoded, nil
			}
		}
		// DO NOT fall back to conn.SFTP.Getwd() which always returns "/" on WSL!
		return "", fmt.Errorf("could not determine terminal CWD for WSL")
	}

	// For standard/other remote SSH connections, attempt to query CWD via background SSH session execution.
	if conn.Client != nil {
		decoded, errQuery := queryRemoteCWD(conn.Client)
		if errQuery == nil && decoded != "" && strings.HasPrefix(decoded, "/") {
			return decoded, nil
		}
	}

	pathStr, err := conn.SFTP.Getwd()
	return pathStr, err
}

// queryRemoteCWD executes the /proc-based CWD lookup over a background SSH session on standard remote hosts.
// This is defined as a package-level variable to allow unit tests to mock remote execution safely.
var queryRemoteCWD = func(client interfaces.SSHClient) (string, error) {
	sess, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	stdout, errPipe := sess.StdoutPipe()
	if errPipe != nil {
		return "", errPipe
	}
	errRun := sess.Run(getCWDScript)
	if errRun != nil {
		return "", errRun
	}
	outputBytes, errRead := io.ReadAll(stdout)
	if errRead != nil {
		return "", errRead
	}
	return strings.TrimSpace(string(outputBytes)), nil
}
