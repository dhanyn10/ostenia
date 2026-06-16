package backend

import (
	"encoding/json"
	"os"
	"ostenia/internal/config"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ProfileData represents the structure for exporting and importing application profiles
type ProfileData struct {
	Config      *config.Config      `json:"config,omitempty"`
	SSHSessions []config.SSHSession `json:"sshSessions,omitempty"`
}

// ExportProfile exports the application configuration and/or SSH sessions to a JSON file
func (a *App) ExportProfile(includeConfig bool, includeSSH bool) error {
	profile := ProfileData{}
	if includeConfig {
		profile.Config = a.cfg
	}
	if includeSSH {
		sessions, err := config.LoadSSHSessions()
		if err == nil {
			profile.SSHSessions = sessions
		}
	}

	filePath, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{
		Title:           "Export Ostenia Profile",
		DefaultFilename: "ostenia_profile.json",
		Filters: []wruntime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil || filePath == "" {
		return err
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// ImportProfile imports the application configuration and/or SSH sessions from a JSON file
func (a *App) ImportProfile() error {
	filePath, err := wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "Import Ostenia Profile",
		Filters: []wruntime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
		},
	})
	if err != nil || filePath == "" {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var profile ProfileData
	err = json.Unmarshal(data, &profile)
	if err != nil {
		return err
	}

	if profile.Config != nil {
		// We preserve BaseDir and WWWRoot to avoid breaking the current installation
		profile.Config.BaseDir = a.cfg.BaseDir
		profile.Config.WWWRoot = a.cfg.WWWRoot
		a.cfg = profile.Config
		_ = config.SaveConfig(a.cfg)
	}

	if profile.SSHSessions != nil {
		_ = config.SaveSSHSessions(profile.SSHSessions)
	}

	wruntime.EventsEmit(a.ctx, "environment_changed", a.cfg)
	return nil
}
