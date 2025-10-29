package main

import (
	"fmt"
	"github.com/charmbracelet/bubbletea"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"io"
	"k8s.io/klog/v2"
	"os"
)

func main() {
	klog.SetOutput(io.Discard)

	cfg := ui.ParseFlags()
	m := ui.NewMultiClusterModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Sprintf("Panic recovered: %v", r))
		}
		logger.Info("Application shutting down")
		logger.Close()
	}()
	logger.Info("Starting Bubbletea program")
	if _, err := p.Run(); err != nil {
		logger.Error(fmt.Sprintf("Bubbletea program error: %v", err))
		os.Exit(1)
	}
	logger.Info("Application completed successfully")
}
