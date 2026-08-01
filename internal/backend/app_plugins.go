package backend

import (
	"context"
	"fmt"
	"os"
	"ostenia/internal/config"
	"ostenia/internal/plugins"
	"ostenia/internal/plugins/php"
	"ostenia/internal/plugins/python"
	plugins_utils "ostenia/internal/plugins/utils"
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
	err := a.downloader.DownloadAndExtract(a.ctx, task)
	if err == nil {
		_, _, currentPath := a.getPluginPaths(task.Name)

		if task.Name == "PHP" {
			_ = service.UpdatePHPPath(currentPath, true)
		} else if task.Name == "Python" {
			_ = service.UpdatePythonPath(currentPath, true)
		}
		a.orchestrator.RequestRefresh()
	}
	return err
}

// CancelDownload cancels an ongoing plugin download
func (a *App) CancelDownload(taskName string) { a.downloader.CancelDownload(taskName) }

// InstallPluginModule installs a sub-module for a parent plugin (e.g., Composer for PHP)
func (a *App) InstallPluginModule(parentName, moduleName string) error {
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
			_ = service.UpdatePHPPath(currentPath, true)
		}
	case "Python":
		err = python.InstallModule(a.ctx, a.downloader, moduleName, currentPath, emitProgress)
		if err == nil {
			_ = service.UpdatePythonPath(currentPath, true)
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
func (a *App) UninstallPluginModule(parentName, moduleName string) error {
	_, _, currentPath := a.getPluginPaths(parentName)

	var err error
	switch parentName {
	case "PHP":
		err = php.UninstallModule(moduleName, currentPath)
		if err == nil {
			_ = service.UpdatePHPPath(currentPath, true)
		}
	case "Python":
		err = python.UninstallModule(moduleName, currentPath)
		if err == nil {
			_ = service.UpdatePythonPath(currentPath, true)
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
func (a *App) SwitchServiceVersion(serviceName, version string) error {
	category, binDir, currentPath := a.getPluginPaths(serviceName)
	prefix := plugins_utils.GetVersionPrefix(category)
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
		cmdPath := filepath.Join(plugins_utils.GetSystemDirectory(), "cmd.exe")
		cmd := plugins_utils.Executor.Command(cmdPath, "/c", "mklink", "/J", currentPath, targetDir)
		cmd.Env = plugins_utils.SafeEnv()
		plugins_utils.SetHideWindow(cmd)
		_ = cmd.Run()
	}
	if category == "php" {
		_ = os.Setenv("PATH", currentPath+";"+os.Getenv("PATH")) // NOSONAR
		_ = service.UpdatePHPConfig(currentPath)
		_ = service.UpdatePHPPath(currentPath, true)
	}
	if category == "nodejs" {
		_ = service.UpdateNodePath(currentPath, true)
	}
	if category == "python" {
		_ = service.UpdatePythonPath(currentPath, true)
	}
	if wasRunning {
		return a.StartService(serviceName)
	}
	a.orchestrator.RequestRefresh()
	return nil
}

// DeleteVersion deletes a specific version folder of a plugin
func (a *App) DeleteVersion(serviceName, version string) error {
	return a.downloader.DeleteVersion(serviceName, version)
}

func (a *App) getServiceTargetDir(category, binDir string) string {
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

// CopyDir recursively copies a directory tree from src to dst.
func CopyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := plugins_utils.CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *App) validateExecutable(category, targetDir string) error {
	var checkFile string
	switch category {
	case "php":
		checkFile = "php.exe"
	case "apache":
		checkFile = "bin/httpd.exe"
	case "mysql":
		checkFile = "bin/mysqld.exe"
	case "nginx":
		checkFile = "nginx.exe"
	case "nodejs":
		checkFile = "node.exe"
	case "python":
		checkFile = "python.exe"
	}

	if checkFile != "" {
		cf := filepath.Join(targetDir, checkFile)
		if category == "apache" {
			if _, err := os.Stat(cf); os.IsNotExist(err) {
				cf = filepath.Join(targetDir, "Apache24", "bin", "httpd.exe")
			}
		}

		if _, err := os.Stat(cf); os.IsNotExist(err) {
			return fmt.Errorf("invalid folder structure: executable (%s) not found in the custom folder", checkFile)
		}
	}
	return nil
}

func (a *App) extractAndProcessZip(serviceName, category, binDir, zipFilePath, targetName string) error {
	targetDir := filepath.Join(binDir, targetName)
	extractTmp := targetDir + ".tmp"
	_ = os.RemoveAll(extractTmp)

	emitProgress := func(ctx context.Context, name string, optionalData ...interface{}) {
		// no-op
	}

	if err := plugins.Unzip(a.ctx, zipFilePath, extractTmp, serviceName, emitProgress); err != nil {
		return fmt.Errorf("failed to extract ZIP: %w", err)
	}

	var err error
	if mgr, ok := a.downloader.(*plugins.Manager); ok {
		err = mgr.PostProcessExtractionManual(extractTmp, targetDir)
	} else {
		_ = os.RemoveAll(targetDir)
		err = os.Rename(extractTmp, targetDir)
	}
	if err != nil {
		_ = os.RemoveAll(extractTmp)
		return fmt.Errorf("failed to post-process extraction: %w", err)
	}

	if err := a.validateExecutable(category, targetDir); err != nil {
		_ = os.RemoveAll(targetDir)
		return err
	}
	return nil
}

// ProcessCustomVersion extracts custom plugin archive or copies direct folder
func (a *App) ProcessCustomVersion(serviceName, sourcePath string) error {
	category, binDir, _ := a.getPluginPaths(serviceName)

	info, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	baseName := filepath.Base(sourcePath)
	targetName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	targetDir := filepath.Join(binDir, targetName)

	if info.IsDir() {
		if err := os.MkdirAll(binDir, 0755); err != nil {
			return err
		}
		if filepath.Clean(sourcePath) != filepath.Clean(targetDir) {
			_ = os.RemoveAll(targetDir)
			if err := CopyDir(sourcePath, targetDir); err != nil {
				return fmt.Errorf("failed to copy directory: %w", err)
			}
		}
		if err := a.validateExecutable(category, targetDir); err != nil {
			_ = os.RemoveAll(targetDir)
			return err
		}
	} else {
		ext := strings.ToLower(filepath.Ext(sourcePath))
		if ext != ".zip" && ext != ".nupkg" {
			return fmt.Errorf("unsupported file format. Please drop/select a .zip or .nupkg file")
		}
		if err := a.extractAndProcessZip(serviceName, category, binDir, sourcePath, targetName); err != nil {
			return err
		}
	}

	a.orchestrator.RequestRefresh()
	return nil
}

// ProcessCustomVersionBytes receives zip bytes from frontend and processes them
func (a *App) ProcessCustomVersionBytes(serviceName, fileName string, fileBytes []byte) error {
	category, binDir, _ := a.getPluginPaths(serviceName)

	if !strings.HasSuffix(strings.ToLower(fileName), ".zip") && !strings.HasSuffix(strings.ToLower(fileName), ".nupkg") {
		return fmt.Errorf("unsupported file format. Please drop a .zip or .nupkg file")
	}

	// Save to a temporary file
	tmpFile := filepath.Join(os.TempDir(), "ostenia_dropped_"+fileName) // NOSONAR
	_ = os.Remove(tmpFile)
	if err := os.WriteFile(tmpFile, fileBytes, 0644); err != nil { // NOSONAR
		return fmt.Errorf("failed to save temp file: %w", err)
	}
	defer os.Remove(tmpFile)

	targetName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	if err := a.extractAndProcessZip(serviceName, category, binDir, tmpFile, targetName); err != nil {
		return err
	}

	a.orchestrator.RequestRefresh()
	return nil
}
