package python

import (
	_ "embed"
	"fmt"
	"os"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"regexp"
	"sort"
)

//go:embed python.svg
var iconSVG string

// DetectVersions scans the NuGet API for available Python versions.
func DetectVersions() ([]string, map[string]string) {
	content := utils.FetchContent("https://api.nuget.org/v3-flatcontainer/python/index.json")
	if content == "" {
		// Fallback to a safe version if API is down
		v := "3.13.13"
		return []string{v}, map[string]string{v: fmt.Sprintf("https://api.nuget.org/v3-flatcontainer/python/%s/python.%s.nupkg", v, v)}
	}

	// Match versions like "3.12.10" or "3.13.13" from the JSON array
	re := regexp.MustCompile(`"(3\.(1[0-9])\.\d+)"`)
	matches := re.FindAllStringSubmatch(content, -1)

	latestPatches := make(map[string]string)
	for _, m := range matches {
		fullVer := m[1]
		minorKey := "3." + m[2]

		if existing, ok := latestPatches[minorKey]; !ok || utils.CompareVersions(fullVer, existing) > 0 {
			latestPatches[minorKey] = fullVer
		}
	}

	var versions []string
	urlMap := make(map[string]string)
	for _, fullVer := range latestPatches {
		versions = append(versions, fullVer)
		urlMap[fullVer] = fmt.Sprintf("https://api.nuget.org/v3-flatcontainer/python/%s/python.%s.nupkg", fullVer, fullVer)
	}

	sort.Slice(versions, func(i, j int) bool {
		return utils.CompareVersions(versions[i], versions[j]) > 0
	})

	return versions, urlMap
}

func GetIcon() string {
	return iconSVG
}

func GetModules() []utils.ModuleDefinition {
	return nil
}

func GetModuleVersion(moduleName, pythonPath string) string {
	return ""
}

func GetInfo(pythonPath string) string {
	pythonExe := filepath.Join(pythonPath, "python.exe")
	if _, err := os.Stat(pythonExe); err != nil {
		return ""
	}
	cmd := utils.Executor.Command(pythonExe, "-m", "pip", "--version")
	utils.SetHideWindow(cmd)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	re := regexp.MustCompile(`pip (\d+\.\d+(?:\.\d+)?)`)
	match := re.FindStringSubmatch(string(out))
	if len(match) > 1 {
		return "Pip " + match[1]
	}
	return ""
}

func UninstallModule(moduleName, pythonPath string) error {
	return fmt.Errorf("unknown module: %s", moduleName)
}

func InstallModule(ctx, m interface{}, moduleName, pythonPath string, emitProgress func(string, float64, string)) error {
	return fmt.Errorf("unknown module: %s", moduleName)
}
