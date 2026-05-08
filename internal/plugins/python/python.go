package python

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"syscall"
)

// DetectVersions scans the Python FTP server for available versions
// and returns only the latest patch for each minor version starting from 3.10,
// verifying that the embeddable package URL actually exists.
func DetectVersions() ([]string, map[string]string) {
	arch := utils.GetSystemArch()
	if arch == "x64" { arch = "amd64" } else { arch = "win32" }

	content := fetchContent("https://www.python.org/ftp/python/")
	re := regexp.MustCompile(`(\d+\.\d+\.\d+)/`)
	matches := re.FindAllStringSubmatch(content, -1)

	latestPatches := make(map[string]int)
	for _, m := range matches {
		v := m[1]
		var major, minor, patch int
		_, err := fmt.Sscanf(v, "%d.%d.%d", &major, &minor, &patch)
		if err != nil { continue }

		// Boundary: Only Python 3.10+
		if major == 3 && minor >= 10 {
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
		expectedURL := fmt.Sprintf("https://www.python.org/ftp/python/%s/python-%s-embed-%s.zip", fullVer, fullVer, arch)

		// Silent check: if doesn't exist, just skip without logging to console
		if checkURLExists(expectedURL) {
			versions = append(versions, fullVer)
			urlMap[fullVer] = expectedURL
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j]) > 0
	})

	if len(versions) == 0 {
		v := "3.12.2"
		fallbackURL := fmt.Sprintf("https://www.python.org/ftp/python/%s/python-%s-embed-%s.zip", v, v, arch)
		return []string{v}, map[string]string{v: fallbackURL}
	}
	return versions, urlMap
}

func checkURLExists(url string) bool {
	resp, err := http.Head(url)
	if err != nil { return false }
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func compareVersions(v1, v2 string) int {
	var a1, b1, c1 int
	var a2, b2, c2 int
	fmt.Sscanf(v1, "%d.%d.%d", &a1, &b1, &c1)
	fmt.Sscanf(v2, "%d.%d.%d", &a2, &b2, &c2)

	if a1 != a2 { return compare(a1, a2) }
	if b1 != b2 { return compare(b1, b2) }
	return compare(c1, c2)
}

func compare(i, j int) int {
	if i > j { return 1 }
	if i < j { return -1 }
	return 0
}

func GetIcon() string {
	data, _ := os.ReadFile(filepath.Join("internal", "plugins", "python", "python.svg"))
	return string(data)
}

func GetModules() []utils.ModuleDefinition {
	return []utils.ModuleDefinition{
		{Name: "Pip", CheckFile: "Scripts/pip.exe"},
	}
}

func GetModuleVersion(moduleName string, pythonPath string) string {
	if moduleName == "Pip" {
		pipExe := filepath.Join(pythonPath, "Scripts", "pip.exe")
		if _, err := os.Stat(pipExe); err != nil { return "" }
		pythonExe := filepath.Join(pythonPath, "python.exe")
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
	if moduleName == "Pip" {
		// Pip is installed into Scripts and Lib/site-packages.
		// For simplicity, we can just remove pip from Scripts
		os.Remove(filepath.Join(pythonPath, "Scripts", "pip.exe"))
		os.Remove(filepath.Join(pythonPath, "Scripts", "pip.exe.manifest"))
		return nil
	}
	return fmt.Errorf("unknown module: %s", moduleName)
}

func InstallModule(ctx interface{}, m interface{}, moduleName string, pythonPath string, emitProgress func(string, float64, string)) error {
	if moduleName == "Pip" {
		emitProgress("Pip", 10, "Downloading get-pip.py...")
		getPipPath := filepath.Join(os.TempDir(), "get-pip.py")
		err := utils.DownloadFile(getPipPath, "https://bootstrap.pypa.io/get-pip.py")
		if err != nil { return err }

		emitProgress("Pip", 50, "Installing Pip...")
		pythonExe := filepath.Join(pythonPath, "python.exe")

		// Note: We're using a simplified command here for modularity.
		// In a real environment, you might want more robust error handling.
		cmd := exec.Command(pythonExe, getPipPath)
		cmd.Dir = pythonPath
		if runtime.GOOS == "windows" {
			cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		}
		err = cmd.Run()
		if err != nil { return err }

		// Patch ._pth file to include site-packages
		files, _ := os.ReadDir(pythonPath)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), "._pth") {
				pthPath := filepath.Join(pythonPath, f.Name())
				content, _ := os.ReadFile(pthPath)
				if !strings.Contains(string(content), "Lib\\site-packages") {
					newContent := string(content) + "\nLib\\site-packages\n"
					os.WriteFile(pthPath, []byte(newContent), 0644)
				}
				break
			}
		}

		emitProgress("Pip", 100, "Completed")
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
