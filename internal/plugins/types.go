// Package plugins provides functionality for managing external components,
// including discovery, downloading, installation, and version management.
package plugins

import "ostenia/internal/backend/interfaces"

// DownloadTask represents a plugin's metadata and state.
type DownloadTask = interfaces.DownloadTask

type PluginModule = interfaces.PluginModule

// Progress reports download/extraction status to frontend.
type Progress = interfaces.Progress
