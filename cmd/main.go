package main

import (
	"log"
	"os"
	"otaviocosta2110/k8s-tui/internal/ui"

	"github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize your main model
	m := ui.NewAppModel()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
		os.Exit(1)
	}
}
