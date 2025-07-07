package main

import (
	"fmt"
	"os"
	"otaviocosta2110/k8s-tui/internal/ui"
	"runtime/debug"

	"github.com/charmbracelet/bubbletea"
)

func main() {
	m := ui.NewAppModel()

	p := tea.NewProgram(m, tea.WithAltScreen())
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			debug.PrintStack()
		}
	}()
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
