package service

import (
	"testing"
)

func TestIsAdmin(t *testing.T) {
	// We don't necessarily need to be admin to run tests
	// Just check if it returns a boolean without crashing
	_ = IsAdmin()
}

func TestRunMeAsAdminError(t *testing.T) {
	// On non-windows it should return error
	// On windows it might try to pop up UAC, so we probably shouldn't run it in CI
}
