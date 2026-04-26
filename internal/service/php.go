package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type PHPExtensionInfo struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
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

	return os.WriteFile(iniPath, []byte(content), 0644)
}

// GetPHPExtensions reads php.ini and returns list of extensions with their status
func GetPHPExtensions(phpPath string) ([]PHPExtensionInfo, error) {
	iniPath := filepath.Join(phpPath, "php.ini")

	// If php.ini doesn't exist, try to initialize it first
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		UpdatePHPConfig(phpPath)
	}

	data, err := os.ReadFile(iniPath)
	if err != nil {
		return nil, fmt.Errorf("could not read php.ini at %s: %v", iniPath, err)
	}

	content := string(data)

	// Even more inclusive regex: matches anything starting with extension= or ;extension=
	// and extracts the name part regardless of prefix/suffix
	re := regexp.MustCompile(`(?m)^;?\s*extension\s*=\s*["']?(?:php_)?([a-z0-9_]+)(?:\.dll)?["']?`)
	matches := re.FindAllStringSubmatch(content, -1)

	// We also need to check the actual lines to see if they are enabled (no semicolon)
	lines := strings.Split(content, "\n")

	extMap := make(map[string]bool)
	var extensions []PHPExtensionInfo

	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if name == "" || name == "ext" { continue }
		if _, exists := extMap[name]; exists { continue }

		// Determine if enabled by looking at the line again
		enabled := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, ";") && strings.Contains(trimmed, "extension") && strings.Contains(trimmed, m[1]) {
				enabled = true
				break
			}
		}

		extMap[name] = true
		extensions = append(extensions, PHPExtensionInfo{
			Name:    name,
			Enabled: enabled,
		})
	}

	fmt.Printf("[PHP] Found %d extensions in %s\n", len(extensions), iniPath)
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
		// Match the specific extension line
		if strings.Contains(trimmed, "extension") && strings.Contains(trimmed, extName) {
			if enable {
				// Remove leading semicolon and space
				lines[i] = strings.TrimPrefix(trimmed, ";")
				lines[i] = strings.TrimSpace(lines[i])
			} else {
				// Add leading semicolon if not there
				if !strings.HasPrefix(trimmed, ";") {
					lines[i] = ";" + trimmed
				}
			}
			found = true
			break
		}
	}

	if !found && enable {
		// If not found, append to end
		lines = append(lines, fmt.Sprintf("extension=%s", extName))
	}

	return os.WriteFile(iniPath, []byte(strings.Join(lines, "\n")), 0644)
}
