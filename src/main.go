package main

import (
	"os"
	"otaviocosta2110/k8s-tui/src/app"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	initialModel := app.NewAppModel()
	
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
