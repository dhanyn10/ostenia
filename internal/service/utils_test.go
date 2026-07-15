package service

import (
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
	"testing"
)

func TestUtils(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()
	utils.Executor = &testutil.MockExecutor{Output: "mocked"}

	t.Run("IsAdmin", func(t *testing.T) {
		// Just call it to see if it doesn't crash
		_ = IsAdmin()
	})

	t.Run("RunMeAsAdmin", func(t *testing.T) {
		_ = RunMeAsAdmin()
	})

	t.Run("ElevateAndExit", func(t *testing.T) {
		// We can't really call ElevateAndExit because it calls os.Exit(0)
	})

	t.Run("AddHostWithElevation", func(t *testing.T) {
		_ = AddHostWithElevation("127.0.0.1", "test.local")
	})
}
