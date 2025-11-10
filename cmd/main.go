package main

import (
	"fmt"
	"github.com/charmbracelet/bubbletea"
	"github.com/otavioCosta2110/k8s-tui/internal/app/cli"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui"
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"io"
	"k8s.io/klog/v2"
	"os"
)

func validateConfig(cfg cli.Config) error {
	// Validate kubeconfig files
	for _, kubeconfig := range cfg.KubeconfigPaths {
		client, err := k8s.NewClient(kubeconfig, cfg.Namespace)
		if err != nil {
			return fmt.Errorf("invalid kubeconfig '%s': %w", kubeconfig, err)
		}
		if client == nil {
			return fmt.Errorf("failed to create Kubernetes client for kubeconfig '%s'", kubeconfig)
		}

		// Validate namespace exists
		namespaces, err := k8s.FetchNamespaces(*client)
		if err != nil {
			return fmt.Errorf("failed to fetch namespaces for kubeconfig '%s': %w", kubeconfig, err)
		}

		// Check if requested namespace exists
		found := false
		for _, ns := range namespaces {
			if ns == cfg.Namespace {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("namespace '%s' does not exist for kubeconfig '%s'. Available namespaces: %v", cfg.Namespace, kubeconfig, namespaces)
		}
	}
	return nil
}

func main() {
	klog.SetOutput(io.Discard)

	cfg := ui.ParseFlags()

	// Validate configuration before starting TUI
	if err := validateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

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
