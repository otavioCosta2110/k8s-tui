package main

import (
	"fmt"
	"os"
	"otaviocosta2110/k8s-tui/src/canvas"
	"otaviocosta2110/k8s-tui/src/kubernetes"

	tea "github.com/charmbracelet/bubbletea"
)

var c *canvas.Canvas

func main() {
  // after implementing cli arguments, make an if to check if there is a kubeconfig already set
  k := kubernetes.NewKubeConfig()
	c = canvas.NewCanvas(k)
	p := tea.NewProgram(c, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}
}
