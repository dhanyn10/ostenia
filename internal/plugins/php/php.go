package php

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"regexp"
	"sort"
)

//go:embed php.svg
var iconSVG string

func DetectVersions() ([]string, map[string]string) {
	baseURL := "https://windows.php.net/downloads/releases/archives/"
	content := utils.FetchContent(baseURL)
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

	sort.Slice(versions, func(i, j int) bool { return utils.CompareVersions(versions[i], versions[j]) > 0 })

	if len(versions) == 0 {
		v := "8.2.12"
		return []string{v}, map[string]string{v: fmt.Sprintf("%sphp-%s-Win32-vs16-%s.zip", baseURL, v, arch)}
	}
	return versions, urlMap
}

func GetIcon() string {
	return iconSVG
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
		utils.SetHideWindow(cmd)
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
		err := utils.DownloadFile(context.Background(), "https://getcomposer.org/composer.phar", composerPhar, "Composer", func(pct float64, status, speed, downloaded string) {
			emitProgress("Composer", pct, status)
		})
		if err != nil { return err }

		batContent := "@php \"%~dp0composer.phar\" %*"
		os.WriteFile(filepath.Join(phpPath, "composer.bat"), []byte(batContent), 0755)

		emitProgress("Composer", 100, "Completed")
		return nil
	}
	return fmt.Errorf("unknown module: %s", moduleName)
}

