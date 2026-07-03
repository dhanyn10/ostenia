package service

import (
	"fmt"
	"os"
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
	cmd := utils.Executor.Command(phpExe, "-v")
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
	if err := initializePHPIni(phpPath, iniPath); err != nil {
		return err
	}

	input, err := os.ReadFile(iniPath)
	if err != nil {
		return err
	}
	content := string(input)

	absPath, _ := filepath.Abs(phpPath)
	extDir := strings.ReplaceAll(filepath.Join(absPath, "ext"), "\\", "/")
	content = configurePHPExtDir(content, extDir)
	content = enableCorePHPExtensions(content)

	return os.WriteFile(iniPath, []byte(content), 0644)
}

func initializePHPIni(phpPath, iniPath string) error {
	if _, err := os.Stat(iniPath); err == nil {
		return nil
	}

	iniDevelopment := filepath.Join(phpPath, "php.ini-development")
	iniProduction := filepath.Join(phpPath, "php.ini-production")

	source := ""
	if _, errDev := os.Stat(iniDevelopment); errDev == nil {
		source = iniDevelopment
	} else if _, errProd := os.Stat(iniProduction); errProd == nil {
		source = iniProduction
	}

	if source != "" {
		data, errRead := os.ReadFile(source)
		if errRead == nil {
			return os.WriteFile(iniPath, data, 0644)
		}
		return errRead
	}
	return nil
}

func configurePHPExtDir(content, extDir string) string {
	reExtDir := regexp.MustCompile(`(?m)^;?\s*extension_dir\s*=\s*".*?"`)
	if reExtDir.MatchString(content) {
		return reExtDir.ReplaceAllString(content, fmt.Sprintf("extension_dir = \"%s\"", extDir))
	}

	reExtDirSimple := regexp.MustCompile(`(?m)^;?\s*extension_dir\s*=\s*ext`)
	if reExtDirSimple.MatchString(content) {
		return reExtDirSimple.ReplaceAllString(content, fmt.Sprintf("extension_dir = \"%s\"", extDir))
	}

	return content + fmt.Sprintf("\nextension_dir = \"%s\"\n", extDir)
}

func enableCorePHPExtensions(content string) string {
	coreExts := []string{"openssl", "mbstring", "curl"}
	for _, ext := range coreExts {
		re := regexp.MustCompile(`(?m)^;\s*(extension\s*=\s*(?:php_)?` + ext + `(?:\.dll)?\s*)$`)
		content = re.ReplaceAllString(content, "$1")
	}
	return content
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
		if name == "" || name == "ext" || extMap[name] {
			continue
		}

		enabled := isPHPExtensionEnabled(lines, m[1])
		extMap[name] = true
		extensions = append(extensions, PHPExtensionInfo{Name: name, Enabled: enabled})
	}
	return extensions, nil
}

func isPHPExtensionEnabled(lines []string, extName string) bool {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, ";") && strings.Contains(trimmed, "extension") && strings.Contains(trimmed, extName) {
			return true
		}
	}
	return false
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
