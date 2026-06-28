package service

import (
	"os"
	"testing"
)

func TestTerminal(t *testing.T) {
	wd, _ := os.Getwd()
	term := NewTerminal(wd, os.Environ())
	if term.WorkingDir != wd {
		t.Errorf("Expected working dir %s, got %s", wd, term.WorkingDir)
	}

	// We won't call Open() as it starts a process
}
