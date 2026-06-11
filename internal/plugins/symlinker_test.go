package plugins

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCreateSymlink(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "symlink_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	targetDir := filepath.Join(tmpDir, "target")
	err = os.Mkdir(targetDir, 0755)
	if err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}

	// Create a dummy file in the target
	dummyFile := filepath.Join(targetDir, "test.txt")
	err = os.WriteFile(dummyFile, []byte("hello"), 0644)
	if err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}

	linkPath := filepath.Join(tmpDir, "link")

	// Test creation
	err = CreateSymlink(targetDir, linkPath)
	if err != nil {
		// On some Windows environments, mklink /J might fail if not permitted
		// but typically it works for junctions without admin.
		t.Fatalf("CreateSymlink() error = %v", err)
	}

	// Verify resolution
	resolved, err := ResolveSymlink(linkPath)
	if err != nil {
		t.Fatalf("ResolveSymlink() error = %v", err)
	}

	// filepath.EvalSymlinks returns the absolute path
	absTarget, _ := filepath.Abs(targetDir)
	absResolved, _ := filepath.Abs(resolved)

	if absResolved != absTarget {
		// On Windows, sometimes paths might have different casing or long/short names
		// but for a junction it should match.
		if runtime.GOOS == "windows" {
			// Junctions might be evaluated slightly differently, let's check if the file exists through the link
			if _, err := os.Stat(filepath.Join(linkPath, "test.txt")); err != nil {
				t.Errorf("Resolved path mismatch: got %v, want %v", absResolved, absTarget)
			}
		} else if absResolved != absTarget {
			t.Errorf("Resolved path mismatch: got %v, want %v", absResolved, absTarget)
		}
	}

	// Test overwriting
	newTargetDir := filepath.Join(tmpDir, "target2")
	_ = os.Mkdir(newTargetDir, 0755)
	err = CreateSymlink(newTargetDir, linkPath)
	if err != nil {
		t.Fatalf("CreateSymlink() failed to overwrite: %v", err)
	}

	resolved, _ = ResolveSymlink(linkPath)
	absNewTarget, _ := filepath.Abs(newTargetDir)
	absResolved, _ = filepath.Abs(resolved)

	if absResolved != absNewTarget {
		if runtime.GOOS == "windows" {
			// Similar check for Windows
			if _, err := os.Stat(filepath.Join(linkPath, "test.txt")); err == nil {
				t.Errorf("Overwrite failed: still points to old target")
			}
		} else {
			t.Errorf("Overwrite mismatch: got %v, want %v", absResolved, absNewTarget)
		}
	}
}
