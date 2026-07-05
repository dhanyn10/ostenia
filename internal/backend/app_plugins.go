package backend

import (
	"fmt"
	"os"
	"ostenia/internal/config"
	"ostenia/internal/plugins"
	"ostenia/internal/plugins/php"
	"ostenia/internal/plugins/python"
	"ostenia/internal/service"
	"path/filepath"
	"strings"
	"time"
)

// GetPrerequisites returns the list of latest known plugin versions for installation
func (a *App) GetPrerequisites() []plugins.DownloadTask { return plugins.GetLatestKnownVersions() }

func (a *App) getPluginPaths(serviceName string) (category string, binDir string, currentPath string) {
	category = strings.ToLower(serviceName)
	if category == "node.js" {
		category = "nodejs"
	}
	binDir = filepath.Join(config.GetBaseDir(), "bin", category)
	currentPath = filepath.Join(binDir, "current")
	return
}

// OpenPluginFolder opens the binary directory for a specific service in File Explorer
func (a *App) OpenPluginFolder(serviceName string) error {
	_, binDir, _ := a.getPluginPaths(serviceName)
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		_ = os.MkdirAll(binDir, 0755)
	}
	return service.OpenExplorer(binDir)
}

// InstallPrerequisite downloads and installs a plugin prerequisite
func (a *App) InstallPrerequisite(task plugins.DownloadTask) error {
	err := a.downloader.DownloadAndExtract(task)
	if err == nil {
		_, _, currentPath := a.getPluginPaths(task.Name)

		if task.Name == "PHP" {
			_ = service.UpdatePHPPath(a.system, currentPath, true)
		} else if task.Name == "Python" {
			_ = service.UpdatePythonPath(a.system, currentPath, true)
		}
		a.orchestrator.RequestRefresh()
	}
	return err
}

// CancelDownload cancels an ongoing plugin download
func (a *App) CancelDownload(taskName string) { a.downloader.CancelDownload(taskName) }

// InstallPluginModule installs a sub-module for a parent plugin (e.g., Composer for PHP)
func (a *App) InstallPluginModule(parentName string, moduleName string) error {
	_, _, currentPath := a.getPluginPaths(parentName)

	if _, err := os.Stat(currentPath); os.IsNotExist(err) {
		return fmt.Errorf("%s is not installed or active", parentName)
	}

	emitProgress := func(name string, pct float64, status string) {
		a.runtime.EventsEmit(a.ctx, "download_progress", plugins.Progress{Name: name, Percentage: pct, Status: status})
	}

	var err error
	switch parentName {
	case "PHP":
		err = php.InstallModule(a.ctx, a.downloader, moduleName, currentPath, emitProgress)
		if err == nil {
			_ = service.UpdatePHPPath(a.system, currentPath, true)
		}
	case "Python":
		err = python.InstallModule(a.ctx, a.downloader, moduleName, currentPath, emitProgress)
		if err == nil {
			_ = service.UpdatePythonPath(a.system, currentPath, true)
		}
	default:
		err = fmt.Errorf("unsupported parent plugin: %s", parentName)
	}

	if err == nil {
		a.orchestrator.RequestRefresh()
	}
	return err
}

// UninstallPluginModule removes a sub-module from a parent plugin
func (a *App) UninstallPluginModule(parentName string, moduleName string) error {
	_, _, currentPath := a.getPluginPaths(parentName)

	var err error
	switch parentName {
	case "PHP":
		err = php.UninstallModule(moduleName, currentPath)
		if err == nil {
			_ = service.UpdatePHPPath(a.system, currentPath, true)
		}
	case "Python":
		err = python.UninstallModule(moduleName, currentPath)
		if err == nil {
			_ = service.UpdatePythonPath(a.system, currentPath, true)
		}
	default:
		err = fmt.Errorf("unsupported parent plugin: %s", parentName)
	}

	if err == nil {
		a.orchestrator.RequestRefresh()
	}
	return err
}

// SwitchServiceVersion changes the active version of a service using directory junctions
func (a *App) SwitchServiceVersion(serviceName string, version string) error {
	category, binDir, currentPath := a.getPluginPaths(serviceName)
	prefix := ""
	switch category {
	case "php":
		prefix = "php-"
	case "apache":
		prefix = "httpd-"
	case "mysql":
		prefix = "mysql-"
	case "nginx":
		prefix = "nginx-"
	case "nodejs":
		prefix = "node-v"
	case "python":
		prefix = "python-"
	}
	targetDir := filepath.Join(binDir, prefix+version)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		targetDir = filepath.Join(binDir, version)
	}
	wasRunning := a.orchestrator.IsRunning(serviceName)
	if wasRunning {
		_ = a.StopService(serviceName)
		time.Sleep(600 * time.Millisecond)
	}
	_ = os.Remove(currentPath)
	if _, err := os.Stat(targetDir); err == nil {
		_ = a.system.CreateSymlink(targetDir, currentPath)
	}
	if category == "php" {
		_ = os.Setenv("PATH", currentPath+";"+os.Getenv("PATH")) // NOSONAR
		_ = service.UpdatePHPConfig(currentPath)
		_ = service.UpdatePHPPath(a.system, currentPath, true)
	}
	if category == "nodejs" {
		_ = service.UpdateNodePath(a.system, currentPath, true)
	}
	if category == "python" {
		_ = service.UpdatePythonPath(a.system, currentPath, true)
	}
	if wasRunning {
		return a.StartService(serviceName)
	}
	a.orchestrator.RequestRefresh()
	return nil
}

// DeleteVersion deletes a specific version folder of a plugin
func (a *App) DeleteVersion(serviceName string, version string) error {
	return a.downloader.DeleteVersion(serviceName, version)
}

func (a *App) getServiceTargetDir(category string, binDir string) string {
	exeMap := map[string]string{
		"nginx":  exeNginx,
		"apache": exeApache,
		"mysql":  exeMySQL,
		"nodejs": exeNode,
		"python": exePython,
	}

	exeName, ok := exeMap[category]
	if !ok {
		return binDir
	}

	var targetDir string
	_ = filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() && info.Name() == exeName {
			targetDir = filepath.Dir(path)
			return filepath.SkipDir
		}
		return nil
	})
	return targetDir
}
