package service

import (
	"testing"
)

func TestPathExistsInString(t *testing.T) {
	pathString := "C:\\bin;C:\\Windows"
	if !pathExistsInString(pathString, "C:\\bin") {
		t.Error("Expected true for C:\\bin")
	}
	if !pathExistsInString(pathString, "C:\\Windows") {
		t.Error("Expected true for C:\\Windows")
	}
	if pathExistsInString(pathString, "C:\\nonexistent") {
		t.Error("Expected false for nonexistent path")
	}
}
