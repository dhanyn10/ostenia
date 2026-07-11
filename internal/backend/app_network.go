package backend

import (
	"fmt"
	"net"
	"os"
	"ostenia/internal/config"
	"path/filepath"
	"time"
)

// ProxyStatusInfo represents the health status of a proxy target
type ProxyStatusInfo struct {
	Name   string `json:"name"`
	IsUp   bool   `json:"isUp"`
	Port   int    `json:"port"`
}

// ProxyAppInfo represents basic information about a potential proxy app directory
type ProxyAppInfo struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}

func (a *App) startProxyWatcher() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			statuses := a.CheckProxyPorts()
			a.runtime.EventsEmit(a.ctx, "proxy_status", statuses)
		}
	}
}

// CheckProxyPorts checks if the configured proxy ports are reachable
func (a *App) CheckProxyPorts() []ProxyStatusInfo {
	var statuses []ProxyStatusInfo
	a.cfgMu.RLock()
	proxies := make(map[string]int)
	if a.cfg != nil {
		for k, v := range a.cfg.Proxies {
			proxies[k] = v
		}
	}
	a.cfgMu.RUnlock()

	for name, port := range proxies {
		isUp := false
		if port > 0 {
			timeout := 500 * time.Millisecond
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), timeout)
			if err == nil {
				isUp = true
				conn.Close()
			}
		}
		statuses = append(statuses, ProxyStatusInfo{Name: name, IsUp: isUp, Port: port})
	}
	return statuses
}

// OpenProxyTerminal opens a terminal at the directory of a proxy app
func (a *App) OpenProxyTerminal(name string, terminalType string) error {
	a.cfgMu.RLock()
	wwwRoot := a.cfg.WWWRoot
	a.cfgMu.RUnlock()
	path := filepath.Join(wwwRoot, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("folder %s not found", name)
	}
	a.OpenTerminalAtPath(terminalType, path)
	return nil
}

// GetProxyApps scans the www directory and returns a list of folders and their configured proxy ports
func (a *App) GetProxyApps() []ProxyAppInfo {
	var apps []ProxyAppInfo
	a.cfgMu.RLock()
	wwwRoot := a.cfg.WWWRoot
	proxies := make(map[string]int)
	if a.cfg != nil {
		for k, v := range a.cfg.Proxies {
			proxies[k] = v
		}
	}
	a.cfgMu.RUnlock()

	entries, err := os.ReadDir(wwwRoot)
	if err != nil {
		return apps
	}

	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			port := 0
			if p, ok := proxies[name]; ok {
				port = p
			}
			apps = append(apps, ProxyAppInfo{Name: name, Port: port})
		}
	}
	return apps
}

// SaveProxyPort saves the proxy port for a specific folder and reconfigures web servers
func (a *App) SaveProxyPort(name string, port int) error {
	a.cfgMu.Lock()
	if a.cfg.Proxies == nil {
		a.cfg.Proxies = make(map[string]int)
	}
	if port <= 0 {
		delete(a.cfg.Proxies, name)
	} else {
		a.cfg.Proxies[name] = port
	}
	cfg := a.cfg
	a.cfgMu.Unlock()
	err := config.SaveConfig(cfg)
	if err != nil {
		return err
	}

	// Trigger web server re-config
	if a.orchestrator.IsRunning("Apache") {
		_ = a.updateApacheConfig(filepath.Join(config.GetBaseDir(), "bin", "apache", "current"), a.orchestrator.GetDetailedInfo("Apache").Port)
		_ = a.StopService("Apache")
		time.Sleep(500 * time.Millisecond)
		_ = a.StartService("Apache")
	}
	if a.orchestrator.IsRunning("Nginx") {
		_ = a.updateNginxConfig(filepath.Join(config.GetBaseDir(), "bin", "nginx", "current"), a.orchestrator.GetDetailedInfo("Nginx").Port)
		_ = a.StopService("Nginx")
		time.Sleep(500 * time.Millisecond)
		_ = a.StartService("Nginx")
	}

	return nil
}
