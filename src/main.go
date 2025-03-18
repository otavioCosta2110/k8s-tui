package main

import (
	"fmt"
	"os"
	"otaviocosta2110/k8s-tui/src/canvas"

	tea "github.com/charmbracelet/bubbletea"
)


func main() {
	p := tea.NewProgram(canvas.Canvas{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}

