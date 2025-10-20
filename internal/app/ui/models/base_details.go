package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"

	tea "github.com/charmbracelet/bubbletea"
)

type baseDetailsModel struct {
	k8sClient  *k8s.Client
	yamlViewer *components.YAMLViewer
	yamlEditor *components.YAMLEditor
	spinner    components.SpinnerModel
	loading    bool
	err        error
	editing    bool
	title      string
}

func newBaseDetailsModel(title, loadingText string) baseDetailsModel {
	return baseDetailsModel{
		yamlViewer: components.NewYAMLViewer(title, "Loading..."),
		spinner:    components.NewSpinner(loadingText),
		loading:    true,
		err:        nil,
		editing:    false,
		title:      title,
	}
}

func (b *baseDetailsModel) setClient(client *k8s.Client) {
	b.k8sClient = client
}

func (b *baseDetailsModel) initCommon(fetchCmd tea.Cmd) tea.Cmd {
	return tea.Batch(b.yamlViewer.Init(), b.spinner.Init(), fetchCmd)
}

func (b *baseDetailsModel) updateCommon(msg tea.Msg) (tea.Cmd, bool) {
	switch msg.(type) {
	case components.SaveMsg:
		return nil, true
	case components.CancelMsg:
		b.editing = false
		return nil, true
	}

	if b.editing && b.yamlEditor != nil {
		var cmd tea.Cmd
		updatedEditor, cmd := b.yamlEditor.Update(msg)
		b.yamlEditor = updatedEditor.(*components.YAMLEditor)
		return cmd, false
	} else if b.yamlViewer != nil {
		var cmd tea.Cmd
		updatedViewer, cmd := b.yamlViewer.Update(msg)
		b.yamlViewer = updatedViewer.(*components.YAMLViewer)
		return cmd, false
	}
	return nil, false
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

func (b *baseDetailsModel) startEditing() {
	if b.yamlViewer != nil {
		content := b.yamlViewer.GetContent()
		b.yamlEditor = components.NewYAMLEditor(b.title, content)
		b.editing = true
	}
}

func (b *baseDetailsModel) View() string {
	if b.loading {
		return b.spinner.CenteredScreenView()
	}
	if b.editing && b.yamlEditor != nil {
		return b.yamlEditor.View()
	}
	return b.yamlViewer.View()
}
