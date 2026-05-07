package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"ostenia/internal/config"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type PHPExtensionInfo struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// GetPHPVersion runs php -v and returns the version string
func GetPHPVersion(currentPath string) (string, error) {
	phpExe := filepath.Join(currentPath, "php.exe")
	if _, err := os.Stat(phpExe); os.IsNotExist(err) {
		return "", err
	}
	cmd := exec.Command(phpExe, "-v")
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func ensureCACert() (string, error) {
	baseDir := config.GetBaseDir()
	certDir := filepath.Join(baseDir, "bin", "php")
	certPath := filepath.Join(certDir, "cacert.pem")

	if _, err := os.Stat(certPath); err == nil {
		return certPath, nil
	}

	if err := os.MkdirAll(certDir, 0755); err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get("https://curl.se/ca/cacert.pem")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download cacert.pem: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(certPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return certPath, err
}

// UpdatePHPConfig ensures required extensions and paths are set in php.ini
func UpdatePHPConfig(phpPath string) error {
	iniPath := filepath.Join(phpPath, "php.ini")
	iniDevelopment := filepath.Join(phpPath, "php.ini-development")
	iniProduction := filepath.Join(phpPath, "php.ini-production")

	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		source := ""
		if _, errDev := os.Stat(iniDevelopment); errDev == nil {
			source = iniDevelopment
		} else if _, errProd := os.Stat(iniProduction); errProd == nil {
			source = iniProduction
		}

		if source != "" {
			data, errRead := os.ReadFile(source)
			if errRead == nil {
				os.WriteFile(iniPath, data, 0644)
			}
		}
	}

	input, err := os.ReadFile(iniPath)
	if err != nil { return err }
	content := string(input)

	absPath, _ := filepath.Abs(phpPath)
	extDir := strings.ReplaceAll(filepath.Join(absPath, "ext"), "\\", "/")

	reExtDir := regexp.MustCompile(`(?m)^;?\s*extension_dir\s*=\s*".*?"`)
	if reExtDir.MatchString(content) {
		content = reExtDir.ReplaceAllString(content, fmt.Sprintf("extension_dir = \"%s\"", extDir))
	} else {
		reExtDirSimple := regexp.MustCompile(`(?m)^;?\s*extension_dir\s*=\s*ext`)
		if reExtDirSimple.MatchString(content) {
			content = reExtDirSimple.ReplaceAllString(content, fmt.Sprintf("extension_dir = \"%s\"", extDir))
		} else {
			content += fmt.Sprintf("\nextension_dir = \"%s\"\n", extDir)
		}
	}

	coreExts := []string{"openssl", "mbstring", "curl"}
	for _, ext := range coreExts {
		re := regexp.MustCompile(`(?m)^;\s*(extension\s*=\s*(?:php_)?` + ext + `(?:\.dll)?\s*)$`)
		content = re.ReplaceAllString(content, "$1")
	}

	// Update SSL CA Cert paths
	if caPath, err := ensureCACert(); err == nil {
		caPath = strings.ReplaceAll(caPath, "\\", "/")

		reCurl := regexp.MustCompile(`(?m)^;?\s*curl\.cainfo\s*=\s*.*$`)
		if reCurl.MatchString(content) {
			content = reCurl.ReplaceAllString(content, fmt.Sprintf("curl.cainfo = \"%s\"", caPath))
		} else {
			content += fmt.Sprintf("\ncurl.cainfo = \"%s\"", caPath)
		}

		reOpenSSL := regexp.MustCompile(`(?m)^;?\s*openssl\.cafile\s*=\s*.*$`)
		if reOpenSSL.MatchString(content) {
			content = reOpenSSL.ReplaceAllString(content, fmt.Sprintf("openssl.cafile = \"%s\"", caPath))
		} else {
			content += fmt.Sprintf("\nopenssl.cafile = \"%s\"", caPath)
		}
	}

	return os.WriteFile(iniPath, []byte(content), 0644)
}

// GetPHPExtensions reads php.ini and returns list of extensions with their status
func GetPHPExtensions(phpPath string) ([]PHPExtensionInfo, error) {
	iniPath := filepath.Join(phpPath, "php.ini")
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		UpdatePHPConfig(phpPath)
	}

	data, err := os.ReadFile(iniPath)
	if err != nil { return nil, err }
	content := string(data)

	re := regexp.MustCompile(`(?m)^;?\s*extension\s*=\s*["']?(?:php_)?([a-z0-9_]+)(?:\.dll)?["']?`)
	matches := re.FindAllStringSubmatch(content, -1)
	lines := strings.Split(content, "\n")

	extMap := make(map[string]bool)
	var extensions []PHPExtensionInfo

	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if name == "" || name == "ext" { continue }
		if _, exists := extMap[name]; exists { continue }

		enabled := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, ";") && strings.Contains(trimmed, "extension") && strings.Contains(trimmed, m[1]) {
				enabled = true
				break
			}
		}

		extMap[name] = true
		extensions = append(extensions, PHPExtensionInfo{Name: name, Enabled: enabled})
	}
	return extensions, nil
}

// TogglePHPExtension enables or disables an extension in php.ini
func TogglePHPExtension(phpPath string, extName string, enable bool) error {
	iniPath := filepath.Join(phpPath, "php.ini")
	data, err := os.ReadFile(iniPath)
	if err != nil { return err }

	content := string(data)
	lines := strings.Split(content, "\n")

	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "extension") && strings.Contains(trimmed, extName) {
			if enable {
				lines[i] = strings.TrimPrefix(trimmed, ";")
				lines[i] = strings.TrimSpace(lines[i])
			} else {
				if !strings.HasPrefix(trimmed, ";") { lines[i] = ";" + trimmed }
			}
			found = true
			break
		}
	}
	if !found && enable {
		lines = append(lines, fmt.Sprintf("extension=%s", extName))
	}
	return os.WriteFile(iniPath, []byte(strings.Join(lines, "\n")), 0644)
}
