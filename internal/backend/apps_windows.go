//go:build windows

package backend

import (
	"golang.org/x/sys/windows/registry"
	"strings"
)

func GetInstalledApps() ([]InstalledApp, error) {
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
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, path, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		subkeys, err := k.ReadSubKeyNames(-1)
		k.Close()
		if err != nil {
			continue
		}

		for _, subkey := range subkeys {
			sk, err := registry.OpenKey(registry.LOCAL_MACHINE, path+`\`+subkey, registry.QUERY_VALUE)
			if err != nil {
				continue
			}

			name, _, _ := sk.GetStringValue("DisplayName")
			location, _, _ := sk.GetStringValue("InstallLocation")
			exe, _, _ := sk.GetStringValue("DisplayIcon") // Often contains the main exe path

			sk.Close()

			if name != "" {
				lowerName := strings.ToLower(name)
				isEditor := false
				for _, kw := range keywords {
					if strings.Contains(lowerName, kw) {
						isEditor = true
						break
					}
				}

				if !isEditor {
					continue
				}

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
		}
	}

	var result []InstalledApp
	for name, path := range apps {
		result = append(result, InstalledApp{Name: name, Path: path})
	}

	return result, nil
}
