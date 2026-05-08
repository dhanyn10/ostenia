package php

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"regexp"
	"sort"
	"syscall"
)

func DetectVersions() ([]string, map[string]string) {
	baseURL := "https://windows.php.net/downloads/releases/archives/"
	content := fetchContent(baseURL)
	arch := utils.GetSystemArch()
	re := regexp.MustCompile(`php-(\d+\.\d+\.\d+)-Win32-vs16-` + arch + `\.zip`)
	matches := re.FindAllStringSubmatch(content, -1)

	// Map untuk menyimpan patch terbaru per minor version (misal: "8.2" -> 14)
	latestPatches := make(map[string]int)
	seen := make(map[string]bool)

	for _, m := range matches {
		v := m[1]
		if seen[v] { continue }
		seen[v] = true

		var major, minor, patch int
		fmt.Sscanf(v, "%d.%d.%d", &major, &minor, &patch)

		// Batasan: PHP 8.2 ke atas
		if major > 8 || (major == 8 && minor >= 2) {
			minorKey := fmt.Sprintf("%d.%d", major, minor)
			if patch > latestPatches[minorKey] {
				latestPatches[minorKey] = patch
			}
		}
	}

	var versions []string
	urlMap := make(map[string]string)
	for minorKey, patch := range latestPatches {
		fullVer := fmt.Sprintf("%s.%d", minorKey, patch)
		versions = append(versions, fullVer)
		urlMap[fullVer] = fmt.Sprintf("%sphp-%s-Win32-vs16-%s.zip", baseURL, fullVer, arch)
	}

	sort.Slice(versions, func(i, j int) bool { return compareVersions(versions[i], versions[j]) > 0 })

	if len(versions) == 0 {
		v := "8.2.12"
		return []string{v}, map[string]string{v: fmt.Sprintf("%sphp-%s-Win32-vs16-%s.zip", baseURL, v, arch)}
	}
	return versions, urlMap
}

func compareVersions(v1, v2 string) int {
	var a1, b1, c1 int
	var a2, b2, c2 int
	fmt.Sscanf(v1, "%d.%d.%d", &a1, &b1, &c1)
	fmt.Sscanf(v2, "%d.%d.%d", &a2, &b2, &c2)
	if a1 != a2 { return a1 - a2 }
	if b1 != b2 { return b1 - b2 }
	return c1 - c2
}

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "php", "php.svg"))
	return string(data)
}

func GetModules() []utils.ModuleDefinition {
	return []utils.ModuleDefinition{
		{Name: "Composer", CheckFile: "composer.phar"},
	}
}

func GetModuleVersion(moduleName string, phpPath string) string {
	if moduleName == "Composer" {
		composerPhar := filepath.Join(phpPath, "composer.phar")
		if _, err := os.Stat(composerPhar); err != nil { return "" }
		phpExe := filepath.Join(phpPath, "php.exe")
		cmd := exec.Command(phpExe, composerPhar, "--version")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		out, err := cmd.Output()
		if err != nil { return "" }
		re := regexp.MustCompile(`Composer version (\d+\.\d+\.\d+)`)
		match := re.FindStringSubmatch(string(out))
		if len(match) > 1 { return match[1] }
	}
	return ""
}

func UninstallModule(moduleName string, phpPath string) error {
	if moduleName == "Composer" {
		os.Remove(filepath.Join(phpPath, "composer.phar"))
		os.Remove(filepath.Join(phpPath, "composer.bat"))
		return nil
	}
	return fmt.Errorf("unknown module: %s", moduleName)
}

func InstallModule(ctx interface{}, m interface{}, moduleName string, phpPath string, emitProgress func(string, float64, string)) error {
	if moduleName == "Composer" {
		emitProgress("Composer", 10, "Downloading...")
		composerPhar := filepath.Join(phpPath, "composer.phar")

		// We need a way to download. Let's assume manager provides it or we do it here.
		// To keep it clean, let's use a helper.
		err := utils.DownloadFile(composerPhar, "https://getcomposer.org/composer.phar")
		if err != nil { return err }

		batContent := "@php \"%~dp0composer.phar\" %*"
		os.WriteFile(filepath.Join(phpPath, "composer.bat"), []byte(batContent), 0755)

		emitProgress("Composer", 100, "Completed")
		return nil
	}
	return fmt.Errorf("unknown module: %s", moduleName)
}

func fetchContent(url string) string {
	resp, err := http.Get(url)
	if err != nil { return "" }
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return string(body)
}
