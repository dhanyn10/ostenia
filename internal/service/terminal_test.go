package service

import (
	"ostenia/internal/plugins/utils"
	"testing"
)

func TestTerminal(t *testing.T) {
	origExecutor := utils.Executor
	defer func() { utils.Executor = origExecutor }()
	utils.Executor = &mockExecutor{}

	term := NewTerminal("/tmp", []string{})

	t.Run("Open", func(t *testing.T) {
		_ = term.Open("cmd")
		_ = term.Open("powershell")
		_ = term.Open("gitbash")
	})

	t.Run("Start", func(t *testing.T) {
		_ = term.Start()
	})

	t.Run("OpenExplorer", func(t *testing.T) {
		_ = OpenExplorer("/tmp")
	})
}
