//go:build !windows

package main

type InstalledApp struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (a *App) GetInstalledApps() ([]InstalledApp, error) {
	return []InstalledApp{}, nil
}
