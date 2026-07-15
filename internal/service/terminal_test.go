package service

import (
	"ostenia/internal/plugins/utils"
	"ostenia/internal/testutil"
	"testing"
)

func TestTerminal(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()
	utils.Executor = &testutil.MockExecutor{Output: "mocked"}

	origGOOS := RuntimeGOOS
	defer func() { RuntimeGOOS = origGOOS }()

	term := NewTerminal("/tmp", []string{})

	t.Run("Open Windows", func(t *testing.T) {
		RuntimeGOOS = "windows"
		_ = term.Open("cmd")
		_ = term.Open("powershell")
		_ = term.Open("gitbash")
		_ = term.Open("unknown")
	})

	t.Run("Open Unix", func(t *testing.T) {
		RuntimeGOOS = "linux"
		_ = term.Open("bash")
	})

	t.Run("Start", func(t *testing.T) {
		_ = term.Start()
	})

	t.Run("OpenExplorer Windows", func(t *testing.T) {
		RuntimeGOOS = "windows"
		_ = OpenExplorer("/tmp")
	})

	t.Run("OpenExplorer Darwin", func(t *testing.T) {
		RuntimeGOOS = "darwin"
		_ = OpenExplorer("/tmp")
	})

	t.Run("OpenExplorer Linux", func(t *testing.T) {
		RuntimeGOOS = "linux"
		_ = OpenExplorer("/tmp")
	})
}
