//go:build windows

package backend

import (
	"golang.org/x/sys/windows/registry"
	"strings"
)

type InstalledApp struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (a *App) GetInstalledApps() ([]InstalledApp, error) {
	apps := make(map[string]string)

	paths := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, path := range paths {
		a.scanRegistryPath(path, apps)
	}

	var result []InstalledApp
	for name, path := range apps {
		result = append(result, InstalledApp{Name: name, Path: path})
	}

	return result, nil
}

func (a *App) scanRegistryPath(path string, apps map[string]string) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer k.Close()

	subkeys, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return
	}

	for _, subkey := range subkeys {
		a.processAppSubkey(path, subkey, apps)
	}
}

func (a *App) processAppSubkey(path, subkey string, apps map[string]string) {
	sk, err := registry.OpenKey(registry.LOCAL_MACHINE, path+`\`+subkey, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer sk.Close()

	name, _, _ := sk.GetStringValue("DisplayName")
	if name == "" || !a.isEditorApp(name) {
		return
	}

	location, _, _ := sk.GetStringValue("InstallLocation")
	exe, _, _ := sk.GetStringValue("DisplayIcon")

	appPath := location
	if appPath == "" && exe != "" {
		// Clean up DisplayIcon which might have ",0" at the end
		appPath = strings.Split(exe, ",")[0]
		appPath = strings.Trim(appPath, `"`)
	}

	if appPath != "" {
		apps[name] = appPath
	}
}

func (a *App) isEditorApp(name string) bool {
	keywords := []string{
		"code", "editor", "notepad", "sublime", "studio", "vim",
		"text", "edit", "writer", "atom", "jetbrains", "intellij",
		"pycharm", "webstorm", "phpstorm", "zed", "cursor", "vscodium",
	}

	lowerName := strings.ToLower(name)
	for _, kw := range keywords {
		if strings.Contains(lowerName, kw) {
			return true
		}
	}
	return false
}
