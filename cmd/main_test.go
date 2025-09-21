package main

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	cfg := ui.ParseFlags()
	pluginManager := plugins.NewPluginManager("./plugins")
	m := ui.NewAppModel(cfg, pluginManager)
	if m == nil {
		t.Error("Expected NewAppModel to return a non-nil model")
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if p == nil {
		t.Error("Expected NewProgram to return a non-nil program")
	}
}

func TestPanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Log("Panic recovered as expected:", r)
		}
	}()
	panic("test panic")
}

func TestProgramExitOnError(t *testing.T) {
	if os.Getenv("TEST_EXIT") == "1" {
		os.Exit(1)
	}
}
