package service

import (
	"ostenia/internal/plugins/utils"
	"testing"
)

func TestUtils(t *testing.T) {
	mockSys := NewMockSystem()

	t.Run("IsAdmin", func(t *testing.T) {
		IsAdmin()
	})

	t.Run("AddHostWithElevation", func(t *testing.T) {
		mockSys.IsAdminVal = true
		_ = AddHostWithElevation(mockSys, "127.0.0.1", "test.local")
	})
}

func TestHideWindow(t *testing.T) {
	cmd := utils.Executor.Command("echo", "test")
	utils.SetHideWindow(cmd)
}
