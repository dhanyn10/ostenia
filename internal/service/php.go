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
		source := iniDevelopment
		if _, errDev := os.Stat(iniDevelopment); os.IsNotExist(errDev) {
			source = iniProduction
		}
		if _, errSrc := os.Stat(source); errSrc == nil {
			data, _ := os.ReadFile(source)
			os.WriteFile(iniPath, data, 0644)
		} else {
			os.WriteFile(iniPath, []byte("[PHP]\nextension_dir = \"ext\"\n"), 0644)
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

	// Just enable core required ones by default if they are commented
	coreExts := []string{"openssl", "mbstring", "curl"}
	for _, ext := range coreExts {
		re := regexp.MustCompile(`(?m)^;\s*(extension\s*=\s*(php_)?` + ext + `(\.dll)?\s*)$`)
		content = re.ReplaceAllString(content, "$1")
	}

	return os.WriteFile(iniPath, []byte(content), 0644)
}

// GetPHPExtensions reads php.ini and returns list of extensions with their status
func GetPHPExtensions(phpPath string) ([]PHPExtensionInfo, error) {
	iniPath := filepath.Join(phpPath, "php.ini")
	data, err := os.ReadFile(iniPath)
	if err != nil { return nil, err }

	content := string(data)
	// Match both enabled and commented extensions
	re := regexp.MustCompile(`(?m)^;?\s*(extension\s*=\s*(?:php_)?([a-z0-9_]+)(?:\.dll)?\s*)$`)
	matches := re.FindAllStringSubmatch(content, -1)

	extMap := make(map[string]bool)
	var extensions []PHPExtensionInfo

	for _, m := range matches {
		line := m[0]
		name := m[3]

		// Skip duplicates in ini
		if _, exists := extMap[name]; exists { continue }
		extMap[name] = true

		enabled := !strings.HasPrefix(line, ";")
		extensions = append(extensions, PHPExtensionInfo{
			Name:    name,
			Enabled: enabled,
		})
	}
	return extensions, nil
}

// TogglePHPExtension enables or disables an extension in php.ini
func TogglePHPExtension(phpPath string, extName string, enable bool) error {
	iniPath := filepath.Join(phpPath, "php.ini")
	data, err := os.ReadFile(iniPath)
	if err != nil { return err }

	content := string(data)

	// Create regex to find the extension line (commented or not)
	// Handles: extension=name, extension=php_name.dll, ;extension=name, etc.
	re := regexp.MustCompile(`(?m)^;?\s*(extension\s*=\s*(php_)?` + regexp.QuoteMeta(extName) + `(\.dll)?\s*)$`)

	if enable {
		// Uncomment
		content = re.ReplaceAllString(content, "$1")
	} else {
		// Comment
		content = re.ReplaceAllString(content, ";$1")
	}

	return os.WriteFile(iniPath, []byte(content), 0644)
}
