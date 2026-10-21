package plugins

import "ostenia/internal/plugins/utils"

// GetInstalledVersionPaths returns a map of version string to its absolute path for a given category.
// It delegates to utils.GetInstalledVersionPaths which provides the canonical implementation.
func GetInstalledVersionPaths(baseDir, category, checkFile string) map[string]string {
	return utils.GetInstalledVersionPaths(baseDir, category, checkFile)
}
