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

	keywords := []string{
		"code", "editor", "notepad", "sublime", "studio", "vim",
		"text", "edit", "writer", "atom", "jetbrains", "intellij",
		"pycharm", "webstorm", "phpstorm", "zed", "cursor", "vscodium",
	}

	paths := []string{
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`,
	}

	for _, path := range paths {
		a.scanRegistryPath(path, keywords, apps)
	}

	var result []InstalledApp
	for name, path := range apps {
		result = append(result, InstalledApp{Name: name, Path: path})
	}

	return result, nil
}

func (a *App) scanRegistryPath(path string, keywords []string, apps map[string]string) {
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
		a.scanRegistrySubkey(path, subkey, keywords, apps)
	}
}

func (a *App) scanRegistrySubkey(path, subkey string, keywords []string, apps map[string]string) {
	sk, err := registry.OpenKey(registry.LOCAL_MACHINE, path+`\`+subkey, registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer sk.Close()

	name, _, _ := sk.GetStringValue("DisplayName")
	if name == "" {
		return
	}

	if !a.isEditorApp(name, keywords) {
		return
	}

	location, _, _ := sk.GetStringValue("InstallLocation")
	exe, _, _ := sk.GetStringValue("DisplayIcon")

	appPath := location
	if appPath == "" && exe != "" {
		appPath = strings.Split(exe, ",")[0]
		appPath = strings.Trim(appPath, `"`)
	}

	if appPath != "" {
		apps[name] = appPath
	}
}

func (a *App) isEditorApp(name string, keywords []string) bool {
	lowerName := strings.ToLower(name)
	for _, kw := range keywords {
		if strings.Contains(lowerName, kw) {
			return true
		}
	}
	return false
}
