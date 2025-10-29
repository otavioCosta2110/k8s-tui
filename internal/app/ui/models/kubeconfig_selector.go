package models

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
)

type KubeconfigSelectedMsg struct {
	Path string
}

type KubeconfigSelectorModel struct {
	list *components.ListModel
}

func NewKubeconfigSelectorModel() *KubeconfigSelectorModel {
	kubeDir := getKubeDir()
	files, err := listKubeconfigFiles(kubeDir)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to list kubeconfig files: %v", err))
		files = []string{}
	}

	list := components.NewList(files, "Select Kubeconfig", func(selected string) tea.Msg {
		return KubeconfigSelectedMsg{Path: filepath.Join(kubeDir, selected)}
	})

	return &KubeconfigSelectorModel{
		list: list,
	}
}

func (m *KubeconfigSelectorModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *KubeconfigSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" {
			return nil, nil // Close selector
		}
	}

	updated, cmd := m.list.Update(msg)
	if list, ok := updated.(*components.ListModel); ok {
		m.list = list
	}
	return m, cmd
}

func (m *KubeconfigSelectorModel) View() string {
	return m.list.View()
}

func getKubeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".kube")
}

func listKubeconfigFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	sort.Strings(files)
	return files, nil
}
