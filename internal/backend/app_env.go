package backend

import (
	"fmt"
	"os"
	"ostenia/internal/config"
	"ostenia/internal/service"
	"path/filepath"
	"strings"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) ensureEnvironmentStructure() {
	baseDir := config.GetBaseDir()
	dirs := []string{filepath.Join(baseDir, "bin"), filepath.Join(baseDir, "ssl")}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			_ = os.MkdirAll(dir, 0755)
		}
	}
	if a.cfg != nil && a.cfg.WWWRoot != "" {
		_ = os.MkdirAll(a.cfg.WWWRoot, 0755)
	}
}

// IsAdmin checks if the application is running with administrative privileges
func (a *App) IsAdmin() bool { return service.IsAdmin() }

// SetWWWRoot sets the server root directory (www)
func (a *App) SetWWWRoot(path string) error {
	fmt.Printf("[App] Setting Server Root (www) to: %s\n", path)
	a.cfg.WWWRoot = path
	err := config.SaveConfig(a.cfg)
	if err != nil { return err }
	_ = os.MkdirAll(path, 0755)
	if a.orchestrator.IsRunning("Apache") {
		_ = a.StopService("Apache")
		time.Sleep(500 * time.Millisecond)
		_ = a.StartService("Apache")
	}
	if a.orchestrator.IsRunning("Nginx") {
		_ = a.StopService("Nginx")
		time.Sleep(500 * time.Millisecond)
		_ = a.StartService("Nginx")
	}
	return nil
}

// SetServerRoot changes the base directory for all Ostenia apps and binaries
func (a *App) SetServerRoot(rootPath string) error {
	fmt.Printf("[App] Switching Apps Location to: %s\n", rootPath)
	a.orchestrator.StopAll()
	time.Sleep(1 * time.Second)
	a.cfg.BaseDir = rootPath
	a.cfg.WWWRoot = filepath.Join(rootPath, "www")
	err := config.SaveConfig(a.cfg)
	if err != nil { return err }
	a.ensureEnvironmentStructure()
	a.orchestrator.RequestRefresh()
	a.runtime.EventsEmit(a.ctx, "environment_changed", a.cfg)
	return nil
}

// SelectServerRoot opens a directory dialog to select the Ostenia apps location
func (a *App) SelectServerRoot() (string, error) {
	selectedDir, err := a.runtime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{Title: "Select Ostenia Apps Location"})
	if err != nil { return "", err }
	if selectedDir != "" { err = a.SetServerRoot(selectedDir); if err != nil { return "", err } }
	return selectedDir, nil
}

// SelectWWWRoot opens a directory dialog to select the server root (www)
func (a *App) SelectWWWRoot() (string, error) {
	selectedDir, err := a.runtime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{Title: "Select Server Root (www)"})
	if err != nil { return "", err }
	if selectedDir != "" { err = a.SetWWWRoot(selectedDir); if err != nil { return "", err } }
	return selectedDir, nil
}

// OpenServerRootFolder opens the www directory in File Explorer
func (a *App) OpenServerRootFolder() error { return service.OpenExplorer(a.cfg.WWWRoot) }

// OpenAppsLocationFolder opens the Ostenia base directory in File Explorer
func (a *App) OpenAppsLocationFolder() error { return service.OpenExplorer(config.GetBaseDir()) }

// OpenTerminal opens a terminal at the current server root directory
func (a *App) OpenTerminal(terminalType string) {
	a.OpenTerminalAtPath(terminalType, a.cfg.WWWRoot)
}

// OpenTerminalAtPath opens a terminal at a specific local path with the Ostenia environment variables set
func (a *App) OpenTerminalAtPath(terminalType string, path string) {
	_, _, phpPath := a.getPluginPaths("PHP")
	_, mysqlBinDir, mysqlCurrentPath := a.getPluginPaths("MySQL")
	mysqlPath := filepath.Join(mysqlCurrentPath, "bin")
	// Fallback if current doesn't exist
	if _, err := os.Stat(mysqlPath); os.IsNotExist(err) {
		_ = filepath.Walk(mysqlBinDir, func(p string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() && info.Name() == exeMySQL {
				mysqlPath = filepath.Dir(p)
				return filepath.SkipDir
			}
			return nil
		})
	}
	_, _, nodePath := a.getPluginPaths("Node.js")

	env := os.Environ()
	pathFound := false
	for i, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			env[i] = "PATH=" + phpPath + ";" + mysqlPath + ";" + nodePath + ";" + e[5:] // NOSONAR
			pathFound = true
			break
		}
	}
	if !pathFound {
		env = append(env, "PATH="+phpPath+";"+mysqlPath+";"+nodePath) // NOSONAR
	}
	cmd := service.NewTerminal(path, env)
	cmd.Open(terminalType)
}
