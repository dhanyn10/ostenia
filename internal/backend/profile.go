package backend

import (
	"context"
	"encoding/json"
	"os"
	"ostenia/internal/config"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ProfileData struct {
	Config      *config.Config      `json:"config,omitempty"`
	SSHSessions []config.SSHSession `json:"sshSessions,omitempty"`
}

type ProfileManager struct {
	Ctx    context.Context
	Config *config.Config
}

func (p *ProfileManager) ExportProfile(includeConfig bool, includeSSH bool) error {
	profile := ProfileData{}
	if includeConfig { profile.Config = p.Config }
	if includeSSH {
		sessions, err := config.LoadSSHSessions()
		if err == nil { profile.SSHSessions = sessions }
	}

	filePath, err := wruntime.SaveFileDialog(p.Ctx, wruntime.SaveDialogOptions{
		Title: "Export Ostenia Profile", DefaultFilename: "ostenia_profile.json",
		Filters: []wruntime.FileFilter{{DisplayName: "JSON Files (*.json)", Pattern: "*.json"}},
	})
	if err != nil || filePath == "" { return err }
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil { return err }
	return os.WriteFile(filePath, data, 0644)
}

func (p *ProfileManager) ImportProfile() error {
	filePath, err := wruntime.OpenFileDialog(p.Ctx, wruntime.OpenDialogOptions{
		Title: "Import Ostenia Profile",
		Filters: []wruntime.FileFilter{{DisplayName: "JSON Files (*.json)", Pattern: "*.json"}},
	})
	if err != nil || filePath == "" { return err }
	data, err := os.ReadFile(filePath)
	if err != nil { return err }
	var profile ProfileData
	if err = json.Unmarshal(data, &profile); err != nil { return err }
	if profile.Config != nil {
		profile.Config.BaseDir = p.Config.BaseDir
		profile.Config.WWWRoot = p.Config.WWWRoot
		p.Config = profile.Config
		_ = config.SaveConfig(p.Config)
	}
	if profile.SSHSessions != nil { _ = config.SaveSSHSessions(profile.SSHSessions) }
	wruntime.EventsEmit(p.Ctx, "environment_changed", p.Config)
	return nil
}
