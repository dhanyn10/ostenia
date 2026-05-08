package python

import (
	_ "embed"
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

//go:embed python.svg
var iconSVG string

// DetectVersions scans the NuGet API for available Python versions.
func DetectVersions() ([]string, map[string]string) {
	content := fetchContent("https://api.nuget.org/v3-flatcontainer/python/index.json")
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

		if existing, ok := latestPatches[minorKey]; !ok || compareVersions(fullVer, existing) > 0 {
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
		return compareVersions(versions[i], versions[j]) > 0
	})

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
	return iconSVG
}

func GetModules() []utils.ModuleDefinition {
	return []utils.ModuleDefinition{
		{Name: "Pip", CheckFile: "python.exe"}, // Nuget Python has pip by default
	}
}

func GetModuleVersion(moduleName string, pythonPath string) string {
	if moduleName == "Pip" {
		pythonExe := filepath.Join(pythonPath, "python.exe")
		if _, err := os.Stat(pythonExe); err != nil { return "" }
		cmd := exec.Command(pythonExe, "-m", "pip", "--version")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		out, err := cmd.Output()
		if err != nil { return "" }
		re := regexp.MustCompile(`pip (\d+\.\d+(?:\.\d+)?)`)
		match := re.FindStringSubmatch(string(out))
		if len(match) > 1 { return match[1] }
	}
	return ""
}

func UninstallModule(moduleName string, pythonPath string) error {
	// For Nuget Python, pip is built-in. Uninstalling it might be counter-productive
	// but we can satisfy the interface.
	return nil
}

func InstallModule(ctx interface{}, m interface{}, moduleName string, pythonPath string, emitProgress func(string, float64, string)) error {
	if moduleName == "Pip" {
		emitProgress("Pip", 100, "Ready")
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
