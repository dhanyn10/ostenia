package plugins

import (
	"os"
	"ostenia/internal/plugins/openssl"
	"ostenia/internal/plugins/utils"
	"path/filepath"
	"sort"
)

type pluginDefinition struct {
	Name             string
	Category         string
	TargetPrefix     string
	CheckFile        string
	Detect           func() ([]string, map[string]string)
	GetIcon          func() string
	GetInfo          func(path string) string
	GetModules       func() []utils.ModuleDefinition
	GetModuleVersion func(name string, path string) string
}

var (
	DetectHeidiSQLInstallationOverride func() (string, string)
	unzipFunc                          = Unzip
)

func DetectHeidiSQLInstallation() (string, string) {
	if DetectHeidiSQLInstallationOverride != nil {
		return DetectHeidiSQLInstallationOverride()
	}
	return utils.DetectHeidiSQLInstallation()
}

func GetLatestKnownVersions() []DownloadTask {
	return LoadPluginsFromJSON()
}

func createDownloadTask(def pluginDefinition, baseDir string) DownloadTask {
	vers, urls := def.Detect()
	t := DownloadTask{
		Name:        def.Name,
		CheckFile:   def.CheckFile,
		IconSVG:     def.GetIcon(),
		VersionUrls: urls,
		Versions:    vers,
	}

	if len(vers) > 0 {
		t.Version = vers[0]
		t.URL = urls[vers[0]]
		t.Target = def.TargetPrefix + vers[0]
	}

	// 1. Detect ALL installed versions
	installedMap := utils.GetInstalledVersionPaths(baseDir, def.Category, t.CheckFile)
	t.InstalledVers = make([]string, 0, len(installedMap))
	for v := range installedMap {
		t.InstalledVers = append(t.InstalledVers, v)
	}
	sort.Strings(t.InstalledVers)

	// Handle Special Cases
	if t.Name == "HeidiSQL" {
		handleHeidiSQLDetection(&t)
		return t
	}
	if t.Name == "OpenSSL" {
		handleOpenSSLDetection(&t)
		return t
	}

	// 1.5 Detect Modules and Info
	currentPath := filepath.Join(baseDir, "bin", def.Category, "current")
	detectModules(&t, def, currentPath)

	if def.GetInfo != nil {
		t.Info = def.GetInfo(currentPath)
	}

	// 2. Check if the currently active 'current' link is functional
	if checkCurrentFunctionality(&t, currentPath, baseDir) {
		t.IsInstalled = true
		if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
			activeFolder := filepath.Base(resolved)
			if activeFolder != "current" {
				activeVer := utils.NormalizeVersion(activeFolder)
				if activeVer != "current" {
					t.ActiveVersion = activeVer
				}
			}
		}
	}

	return t
}

func handleHeidiSQLDetection(t *DownloadTask) {
	exePath, _ := DetectHeidiSQLInstallation()
	if exePath != "" {
		t.IsInstalled = true
		if len(t.InstalledVers) == 0 {
			t.InstalledVers = []string{"System"}
		}
	}
}

func handleOpenSSLDetection(t *DownloadTask) {
	t.InstalledVers = nil
	t.IsInstalled = false
	if gv := openssl.DetectInstalledVersion(); gv != "" {
		t.Version = gv
		t.InstalledVers = []string{gv}
		t.IsInstalled = true
	}
}

func detectModules(t *DownloadTask, def pluginDefinition, currentPath string) {
	if def.GetModules == nil {
		return
	}
	for _, modDef := range def.GetModules() {
		isModInstalled := false
		if _, err := os.Stat(filepath.Join(currentPath, modDef.CheckFile)); err == nil {
			isModInstalled = true
		}
		status := "Not Installed"
		version := ""
		if isModInstalled {
			status = "Ready"
			if def.GetModuleVersion != nil {
				version = def.GetModuleVersion(modDef.Name, currentPath)
			}
		}
		t.Modules = append(t.Modules, PluginModule{
			Name: modDef.Name, IsInstalled: isModInstalled, Status: status, Version: version, CheckFile: modDef.CheckFile,
		})
	}
}

func checkCurrentFunctionality(t *DownloadTask, currentPath string, baseDir string) bool {
	if resolved, err := filepath.EvalSymlinks(currentPath); err == nil {
		return checkFileExists(t.Name, resolved, t.CheckFile)
	}

	if t.Target != "" {
		targetPath := filepath.Join(baseDir, "bin", t.Target)
		return checkFileExists(t.Name, targetPath, t.CheckFile)
	}

	return false
}

func checkFileExists(pluginName, basePath, checkFile string) bool {
	cf := filepath.Join(basePath, checkFile)
	if pluginName == "Apache" {
		if _, err := os.Stat(cf); os.IsNotExist(err) {
			cf = filepath.Join(basePath, "Apache24", "bin", "httpd.exe")
		}
	}

	_, err := os.Stat(cf)
	return err == nil
}
