package main

import (
	"fmt"
	"os"
	"otaviocosta2110/k8s-tui/src/canvas"

	tea "github.com/charmbracelet/bubbletea"
)

var c *canvas.Canvas

func main() {
	c = canvas.NewCanvas()
	p := tea.NewProgram(c)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}
