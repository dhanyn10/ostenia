package plugins

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestManagerDeleteVersion(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "mgr_delete_test")
	defer os.RemoveAll(tmpDir)

	// Mocking GetBaseDir via environment
	os.Setenv("OSTENIA_HOME", tmpDir)
	defer os.Unsetenv("OSTENIA_HOME")

	phpDir := filepath.Join(tmpDir, "bin", "php", "php-8.2.0")
	os.MkdirAll(phpDir, 0755)

	mgr := NewManager(context.Background())
	err := mgr.DeleteVersion("php", "8.2.0")
	if err != nil {
		t.Fatalf("DeleteVersion failed: %v", err)
	}

	if _, err := os.Stat(phpDir); !os.IsNotExist(err) {
		t.Error("Directory should have been deleted")
	}
}

func TestManagerCancelDownload(t *testing.T) {
	mgr := NewManager(context.Background())
	ctx, cancel := context.WithCancel(context.Background())
	mgr.cancels["test"] = cancel

	mgr.CancelDownload("test")

	mgr.cancelsMu.Lock()
	_, ok := mgr.cancels["test"]
	mgr.cancelsMu.Unlock()
	if ok {
		t.Error("Cancel function should have been removed from map")
	}

	if ctx.Err() != context.Canceled {
		t.Error("Context should have been canceled")
	}
}
