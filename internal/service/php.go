package service

import (
	"fmt"
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"regexp"
	"strings"
)

// PHPExtensionInfo contains the name and enabled status of a PHP extension
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
	utils.SetHideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// UpdatePHPConfig ensures required extensions and paths are set in php.ini.
// It initializes php.ini from development/production templates if it doesn't exist.
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
				_ = os.WriteFile(iniPath, data, 0644)
			}
		}
	}

	input, err := os.ReadFile(iniPath)
	if err != nil {
		return err
	}
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

	// Enable standard required extensions
	coreExts := []string{"openssl", "mbstring", "curl"}
	for _, ext := range coreExts {
		re := regexp.MustCompile(`(?m)^;\s*(extension\s*=\s*(?:php_)?` + ext + `(?:\.dll)?\s*)$`)
		content = re.ReplaceAllString(content, "$1")
	}

	return os.WriteFile(iniPath, []byte(content), 0644)
}

// GetPHPExtensions reads php.ini and returns list of extensions with their status
func GetPHPExtensions(phpPath string) ([]PHPExtensionInfo, error) {
	iniPath := filepath.Join(phpPath, "php.ini")
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		_ = UpdatePHPConfig(phpPath)
	}

	data, err := os.ReadFile(iniPath)
	if err != nil {
		return nil, err
	}
	content := string(data)

	re := regexp.MustCompile(`(?m)^;?\s*extension\s*=\s*["']?(?:php_)?([a-z0-9_]+)(?:\.dll)?["']?`)
	matches := re.FindAllStringSubmatch(content, -1)
	lines := strings.Split(content, "\n")

	extMap := make(map[string]bool)
	var extensions []PHPExtensionInfo

	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if name == "" || name == "ext" {
			continue
		}
		if _, exists := extMap[name]; exists {
			continue
		}

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
	if err != nil {
		return err
	}

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
				if !strings.HasPrefix(trimmed, ";") {
					lines[i] = ";" + trimmed
				}
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
