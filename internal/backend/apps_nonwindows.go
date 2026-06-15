//go:build !windows

package backend

func GetInstalledApps() ([]InstalledApp, error) {
	return []InstalledApp{}, nil
}
