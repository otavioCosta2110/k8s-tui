package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	k8s "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type deploymentDetailsModel struct {
	baseDetailsModel
	deployment *k8s.DeploymentInfo
	editor     *components.YAMLEditor
	isEditing  bool
}

type deploymentDetailsLoadedMsg struct {
	content string
	err     error
}

func NewDeploymentDetails(k k8s.Client, namespace, deploymentName string) *deploymentDetailsModel {
	return &deploymentDetailsModel{
		baseDetailsModel: newBaseDetailsModel("Deployment: "+deploymentName, "Loading deployment details..."),
		deployment:       k8s.NewDeploymentInfo(deploymentName, namespace, k),
		isEditing:        false,
	}
}

func (d *deploymentDetailsModel) InitComponent(k *k8s.Client) (tea.Model, error) {
	d.setClient(k)
	return d, nil
}

func (d *deploymentDetailsModel) fetchDeploymentDetails() tea.Cmd {
	return func() tea.Msg {
		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*d.k8sClient)
		desc, err := api.DescribeDeployment(d.deployment.Namespace, d.deployment.Name)
		return deploymentDetailsLoadedMsg{content: desc, err: err}
	}
}

func (d *deploymentDetailsModel) Init() tea.Cmd {
	return d.initCommon(d.fetchDeploymentDetails())
}

func (d *deploymentDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case deploymentDetailsLoadedMsg:
		d.loading = false
		if msg.err != nil {
			d.err = msg.err
			d.yamlViewer.SetContent("Error loading deployment details: " + msg.err.Error())
		} else {
			d.yamlViewer.SetContent(msg.content)
		}
		return d, nil
	case components.EditMsg:
		d.isEditing = true
		d.editor = components.NewYAMLEditorWithHelp(
			"Deployment: "+d.deployment.Name,
			msg.Content,
			"Esc: Cancel",
		)
		return d, d.editor.Init()

	case components.SaveMsg:
		err := d.deployment.Apply(msg.Content)
		if err != nil {
			d.err = err
			d.isEditing = false
			d.editor = nil
			return d, nil
		}

		d.isEditing = false
		d.editor = nil

		var desc string

		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*d.k8sClient)
		desc, err = api.DescribeDeployment(d.deployment.Namespace, d.deployment.Name)

		if err != nil {
			d.err = err
			return d, nil
		}

		d.yamlViewer = components.NewYAMLViewerWithHelp(
			"Deployment: "+d.deployment.Name,
			desc,
			"↑/↓: Scroll • e: Edit • q: Quit",
		)
		return d, d.yamlViewer.Init()

	case components.CancelMsg:
		d.isEditing = false
		d.editor = nil
		return d, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return d, tea.Quit
		}
	}

	if d.isEditing && d.editor != nil {
		updatedModel, cmd := d.editor.Update(msg)
		if editor, ok := updatedModel.(*components.YAMLEditor); ok {
			d.editor = editor
		}
		return d, cmd
	} else if d.yamlViewer != nil {
		updatedModel, cmd := d.yamlViewer.Update(msg)
		if viewer, ok := updatedModel.(*components.YAMLViewer); ok {
			d.yamlViewer = viewer
		}
		var spinnerCmd tea.Cmd
		d.spinner, spinnerCmd = d.spinner.Update(msg)
		return d, tea.Batch(cmd, spinnerCmd)
	}

	return d, nil
}

func (d *deploymentDetailsModel) View() string {
	if d.err != nil {
		return lipgloss.NewStyle().
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render("Error: " + d.err.Error())
	}

	if d.isEditing && d.editor != nil {
		return d.editor.View()
	}

	return d.baseDetailsModel.View()
}
