package backend

import (
	"context"
	"fmt"
	"os"
	"ostenia/internal/config"
	"ostenia/internal/service"
	"path/filepath"
	"strings"
)

type TerminalManager struct {
	Ctx            context.Context
	ServiceManager *ServiceManager
	Cfg            *config.Config
}

func (t *TerminalManager) OpenTerminal(terminalType string) {
	t.OpenTerminalAtPath(terminalType, t.Cfg.WWWRoot)
}

func (t *TerminalManager) OpenTerminalAtPath(terminalType string, path string) {
	_, _, phpPath := t.ServiceManager.GetPluginPaths("PHP")
	_, mysqlBinDir, mysqlCurrentPath := t.ServiceManager.GetPluginPaths("MySQL")
	mysqlPath := filepath.Join(mysqlCurrentPath, "bin")

	if _, err := os.Stat(mysqlPath); os.IsNotExist(err) {
		_ = filepath.Walk(mysqlBinDir, func(p string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() && info.Name() == exeMySQL {
				mysqlPath = filepath.Dir(p)
				return filepath.SkipDir
			}
			return nil
		})
	}
	_, _, nodePath := t.ServiceManager.GetPluginPaths("Node.js")

	env := os.Environ()
	pathFound := false
	for i, e := range env {
		if strings.HasPrefix(strings.ToUpper(e), "PATH=") {
			env[i] = "PATH=" + phpPath + ";" + mysqlPath + ";" + nodePath + ";" + e[5:]
			pathFound = true
			break
		}
	}
	if !pathFound {
		env = append(env, "PATH="+phpPath+";"+mysqlPath+";"+nodePath)
	}
	cmd := service.NewTerminal(path, env)
	cmd.Open(terminalType)
}

func (t *TerminalManager) OpenServiceTerminal(serviceName string, terminalType string) error {
	category, binDir, _ := t.ServiceManager.GetPluginPaths(serviceName)
	targetDir := t.ServiceManager.GetServiceTargetDir(category, binDir)

	if targetDir == "" {
		targetDir = binDir
	}
	t.OpenTerminalAtPath(terminalType, targetDir)
	return nil
}

func (t *TerminalManager) OpenProxyTerminal(name string, terminalType string) error {
	path := filepath.Join(t.Cfg.WWWRoot, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("folder %s not found", name)
	}
	t.OpenTerminalAtPath(terminalType, path)
	return nil
}
