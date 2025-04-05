package main

import (
	"fmt"
	"os"
	"otaviocosta2110/k8s-tui/src/app"
	"otaviocosta2110/k8s-tui/src/kubernetes"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	k := kubernetes.NewKubeConfig()
	initialModel := app.NewAppModel(k)
	
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}
