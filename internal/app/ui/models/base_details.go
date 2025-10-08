package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"

	tea "github.com/charmbracelet/bubbletea"
)

type baseDetailsModel struct {
	k8sClient  *k8s.Client
	yamlViewer *components.YAMLViewer
	spinner    components.SpinnerModel
	loading    bool
	err        error
}

func newBaseDetailsModel(title, loadingText string) baseDetailsModel {
	return baseDetailsModel{
		yamlViewer: components.NewYAMLViewer(title, "Loading..."),
		spinner:    components.NewSpinner(loadingText),
		loading:    true,
		err:        nil,
	}
}

func (b *baseDetailsModel) setClient(client *k8s.Client) {
	b.k8sClient = client
}

func (b *baseDetailsModel) initCommon(fetchCmd tea.Cmd) tea.Cmd {
	return tea.Batch(b.yamlViewer.Init(), b.spinner.Init(), fetchCmd)
}

func (b *baseDetailsModel) updateCommon(msg tea.Msg) (tea.Cmd, bool) {
	var cmd tea.Cmd
	updatedViewer, cmd := b.yamlViewer.Update(msg)
	b.yamlViewer = updatedViewer.(*components.YAMLViewer)

	var spinnerCmd tea.Cmd
	b.spinner, spinnerCmd = b.spinner.Update(msg)

	return tea.Batch(cmd, spinnerCmd), false
}

func (b *baseDetailsModel) setLoadedContent(content string, err error) {
	b.loading = false
	if err != nil {
		b.err = err
		b.yamlViewer.SetContent("Error loading details: " + err.Error())
	} else {
		b.yamlViewer.SetContent(content)
	}
}

func (b *baseDetailsModel) View() string {
	if b.loading {
		return b.spinner.CenteredScreenView()
	}
	return b.yamlViewer.View()
}
