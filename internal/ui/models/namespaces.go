package models

import (
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	ui "github.com/otavioCosta2110/k8s-tui/internal/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/ui/custom_styles"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type namespacesModel struct {
	list      []string
	k8sClient *k8s.Client
	loading   bool
	err       error
}

func NewNamespaces(k k8s.Client) (*namespacesModel, error) {
	namespaces, err := k8s.FetchNamespaces(k)
	if err != nil {
		return nil, err
	}

	return &namespacesModel{
		list:      namespaces,
		k8sClient: &k,
		loading:   false,
		err:       nil,
	}, nil
}

func (n *namespacesModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	n.k8sClient = k
	namespaces, err := k8s.FetchNamespaces(*k)
	if err != nil {
		return nil, err
	}

	var listItems []ui.ListItem
	for _, namespace := range namespaces {
		listItems = append(listItems, ui.NewItem(customstyles.ResourceIcons["Namespaces"]+" "+namespace, ""))
	}

	onSelect := func(selected string) tea.Msg {
		namespace := selected
		if icon, exists := customstyles.ResourceIcons["Namespaces"]; exists && strings.HasPrefix(selected, icon+" ") {
			namespace = strings.TrimPrefix(selected, icon+" ")
		}

		return components.NavigateMsg{
			NewScreen: NewResource(*k, namespace).InitComponent(*k),
		}
	}

	return ui.NewListWithItems(listItems, customstyles.ResourceIcons["Namespaces"]+" Namespaces", onSelect), nil
}
