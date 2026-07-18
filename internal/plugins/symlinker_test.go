package plugins

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func setupSymlinkTest(t *testing.T) (string, string, string) {
	tmpDir, err := os.MkdirTemp("", "symlink_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	targetDir := filepath.Join(tmpDir, "target")
	if err := os.Mkdir(targetDir, 0755); err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "test.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}
	return tmpDir, targetDir, filepath.Join(tmpDir, "link")
}

func TestCreateSymlink_Creation(t *testing.T) {
	tmpDir, targetDir, linkPath := setupSymlinkTest(t)
	defer os.RemoveAll(tmpDir)

	if err := CreateSymlink(targetDir, linkPath); err != nil {
		t.Fatalf("CreateSymlink() error = %v", err)
	}

	resolved, err := ResolveSymlink(linkPath)
	if err != nil {
		t.Fatalf("ResolveSymlink() error = %v", err)
	}
	absTarget, _ := filepath.Abs(targetDir)
	absResolved, _ := filepath.Abs(resolved)

	if absResolved != absTarget {
		if runtime.GOOS == "windows" {
			if _, err := os.Stat(filepath.Join(linkPath, "test.txt")); err != nil {
				t.Errorf("Resolved path mismatch: got %v, want %v", absResolved, absTarget)
			}
		} else {
			t.Errorf("Resolved path mismatch: got %v, want %v", absResolved, absTarget)
		}
	}
}

func TestCreateSymlink_Overwriting(t *testing.T) {
	tmpDir, targetDir, linkPath := setupSymlinkTest(t)
	defer os.RemoveAll(tmpDir)

	if err := CreateSymlink(targetDir, linkPath); err != nil {
		t.Fatalf("CreateSymlink() error = %v", err)
	}

	newTargetDir := filepath.Join(tmpDir, "target2")
	if err := os.Mkdir(newTargetDir, 0755); err != nil {
		t.Fatalf("failed to create new target dir: %v", err)
	}
	if err := CreateSymlink(newTargetDir, linkPath); err != nil {
		t.Fatalf("CreateSymlink() failed to overwrite: %v", err)
	}

	resolved, err := ResolveSymlink(linkPath)
	if err != nil {
		t.Fatalf("ResolveSymlink() error = %v", err)
	}
	absNewTarget, _ := filepath.Abs(newTargetDir)
	absResolved, _ := filepath.Abs(resolved)

	if absResolved != absNewTarget {
		if runtime.GOOS == "windows" {
			if _, err := os.Stat(filepath.Join(linkPath, "test.txt")); err == nil {
				t.Errorf("Overwrite failed: still points to old target")
			}
		} else {
			t.Errorf("Overwrite mismatch: got %v, want %v", absResolved, absNewTarget)
		}
	}
}
