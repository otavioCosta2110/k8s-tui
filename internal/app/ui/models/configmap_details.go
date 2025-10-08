package models

import (
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/components"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
	resources "github.com/otavioCosta2110/k8s-tui/internal/k8s/resources"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type cmDetailsModel struct {
	cm         *resources.Configmap
	k8sClient  *resources.Client
	loading    bool
	err        error
	yamlViewer *components.YAMLViewer
	editor     *components.YAMLEditor
	spinner    components.SpinnerModel
	isEditing  bool
}

type cmDetailsLoadedMsg struct {
	content string
	err     error
}

func NewConfigmapDetails(k resources.Client, namespace, cmName string) *cmDetailsModel {
	return &cmDetailsModel{
		cm:         resources.NewConfigmap(cmName, namespace, k),
		k8sClient:  &k,
		yamlViewer: components.NewYAMLViewerWithHelp("Configmap: "+cmName, "Loading...", "↑/↓: Scroll • e: Edit • q: Quit"),
		spinner:    components.NewSpinner("Loading configmap details..."),
		loading:    true,
		err:        nil,
		isEditing:  false,
	}
}

func (c *cmDetailsModel) InitComponent(k *resources.Client) (tea.Model, error) {
	c.k8sClient = k
	return c, nil
}

func (c *cmDetailsModel) fetchConfigmapDetails() tea.Cmd {
	return func() tea.Msg {
		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*c.k8sClient)
		desc, err := api.DescribeConfigMap(c.cm.Namespace, c.cm.Name)
		return cmDetailsLoadedMsg{content: desc, err: err}
	}
}

func (c *cmDetailsModel) Init() tea.Cmd {
	return tea.Batch(c.yamlViewer.Init(), c.spinner.Init(), c.fetchConfigmapDetails())
}

func (c *cmDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cmDetailsLoadedMsg:
		c.loading = false
		if msg.err != nil {
			c.err = msg.err
			c.yamlViewer.SetContent("Error loading configmap details: " + msg.err.Error())
		} else {
			c.yamlViewer.SetContent(msg.content)
		}
		return c, nil
	case components.EditMsg:
		c.isEditing = true
		c.editor = components.NewYAMLEditorWithHelp(
			"Configmap: "+c.cm.Name,
			msg.Content,
			"Esc: Cancel",
		)
		return c, c.editor.Init()

	case components.SaveMsg:
		err := c.cm.Update(msg.Content)
		if err != nil {
			c.err = err
			c.isEditing = false
			c.editor = nil
			return c, nil
		}

		c.isEditing = false
		c.editor = nil

		var desc string

		pm := plugins.GetGlobalPluginManager()
		api := pm.GetAPI()
		api.SetClient(*c.k8sClient)
		desc, err = api.DescribeConfigMap(c.cm.Namespace, c.cm.Name)

		if err != nil {
			c.err = err
			return c, nil
		}

		c.yamlViewer = components.NewYAMLViewerWithHelp(
			"Configmap: "+c.cm.Name,
			desc,
			"↑/↓: Scroll • e: Edit • q: Quit",
		)
		return c, c.yamlViewer.Init()

	case components.CancelMsg:
		c.isEditing = false
		c.editor = nil
		return c, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return c, tea.Quit
		}
	}

	if c.isEditing && c.editor != nil {
		updatedModel, cmd := c.editor.Update(msg)
		if editor, ok := updatedModel.(*components.YAMLEditor); ok {
			c.editor = editor
		}
		return c, cmd
	} else if c.yamlViewer != nil {
		updatedModel, cmd := c.yamlViewer.Update(msg)
		if viewer, ok := updatedModel.(*components.YAMLViewer); ok {
			c.yamlViewer = viewer
		}
		var spinnerCmd tea.Cmd
		c.spinner, spinnerCmd = c.spinner.Update(msg)
		return c, tea.Batch(cmd, spinnerCmd)
	}

	return c, nil
}

func (c *cmDetailsModel) View() string {
	if c.err != nil {
		return lipgloss.NewStyle().
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render("Error: " + c.err.Error())
	}

	if c.isEditing && c.editor != nil {
		return c.editor.View()
	}

	if c.loading {
		return c.spinner.View()
	}

	if c.yamlViewer != nil {
		return c.yamlViewer.View()
	}

	return lipgloss.NewStyle().
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render("Loading...")
}
