package models

import (
	// "context"
	"fmt"
	"otaviocosta2110/k8s-tui/internal/k8s"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type podItem struct {
	name      string
	namespace string
	status    string
}

func (i podItem) Title() string       { return i.name }
func (i podItem) Description() string { return fmt.Sprintf("%s | %s", i.namespace, i.status) }
func (i podItem) FilterValue() string { return i.name }

type podsModel struct {
	list     list.Model
	k8sClient *k8s.Client
	loading  bool
	err      error
}

func newPodsModel(client *k8s.Client) podsModel {
	items := []list.Item{} // empty initially
	
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Pods"
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("62"))
	
	return podsModel{
		list:     l,
		k8sClient: client,
		loading:  true,
	}
}

// func (m podsModel) Init() tea.Cmd {
// 	return m.fetchPods
// }

// func (m podsModel) fetchPods() tea.Msg {
// 	pods, err := m.k8sClient.GetPods(context.Background(), "default")
// 	if err != nil {
// 		return errMsg{err}
// 	}
// 	
// 	var items []list.Item
// 	for _, p := range pods {
// 		items = append(items, podItem{
// 			name:      p.Name,
// 			namespace: p.Namespace,
// 			status:    p.Status,
// 		})
// 	}
// 	
// 	return podsMsg(items)
// }

func (m podsModel) Update(msg tea.Msg) (podsModel, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2) // leave space for header/footer
		
	case podsMsg:
		m.loading = false
		m.list.SetItems(msg)
		
	case errMsg:
		m.err = msg
		m.loading = false
		return m, nil
	}
	
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m podsModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}
	
	if m.loading {
		return "Loading pods..."
	}
	
	return m.list.View()
}

// Messages
type podsMsg []list.Item
type errMsg struct{ error }
