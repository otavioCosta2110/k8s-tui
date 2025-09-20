package components

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otavioCosta2110/k8s-tui/pkg/ui/custom_styles"
	"github.com/otavioCosta2110/k8s-tui/pkg/utils"
)

type YAMLEditor struct {
	textarea    textarea.Model
	viewport    viewport.Model
	title       string
	original    string
	ready       bool
	styles      *YAMLViewerStyles
	customHelp  string
	useExternal bool
	tempFile    string
	editorCmd   *exec.Cmd
}

func NewYAMLEditor(title, content string) *YAMLEditor {
	editor := utils.GetPreferredEditor()

	useExternal := editor != "vi"

	if useExternal {
		return NewExternalYAMLEditor(title, content)
	}

	ta := textarea.New()
	ta.SetValue(content)
	ta.Focus()

	return &YAMLEditor{
		textarea:    ta,
		title:       title,
		original:    content,
		styles:      DefaultYAMLViewerStyles,
		useExternal: false,
	}
}

func NewYAMLEditorWithHelp(title, content, helpText string) *YAMLEditor {
	editor := NewYAMLEditor(title, content)
	editor.customHelp = helpText
	return editor
}

func NewExternalYAMLEditor(title, content string) *YAMLEditor {
	tempFile, err := createTempFile(content)
	if err != nil {
		return NewYAMLEditor(title, content)
	}

	return &YAMLEditor{
		title:       title,
		original:    content,
		styles:      DefaultYAMLViewerStyles,
		useExternal: true,
		tempFile:    tempFile,
	}
}

func (m *YAMLEditor) Init() tea.Cmd {
	if m.useExternal {
		return m.launchExternalEditor()
	}
	return textarea.Blink
}

func (m *YAMLEditor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.useExternal {
			switch msg.String() {
			case "esc":
				m.cleanupTempFile()
				return m, func() tea.Msg {
					return CancelMsg{}
				}
			}
		} else {
			switch msg.String() {
			case "ctrl+s":
				return m, func() tea.Msg {
					return SaveMsg{
						Content:    m.textarea.Value(),
						Title:      m.title,
						Original:   m.original,
						IsModified: m.textarea.Value() != m.original,
					}
				}
			case "esc":
				return m, func() tea.Msg {
					return CancelMsg{}
				}
			}
		}
	case ExternalEditorDoneMsg:
		content, err := m.readEditedContent()
		if err != nil {
			return m, func() tea.Msg {
				return ExternalEditorErrorMsg{Error: err}
			}
		}

		m.cleanupTempFile()

		return m, func() tea.Msg {
			return SaveMsg{
				Content:    content,
				Title:      m.title,
				Original:   m.original,
				IsModified: content != m.original,
			}
		}
	case ExternalEditorErrorMsg:
		m.cleanupTempFile()

		errMsg := msg.Error.Error()
		if strings.Contains(errMsg, "cancelled") || strings.Contains(errMsg, "exited with code") {
			return m, func() tea.Msg {
				return CancelMsg{}
			}
		}

		m.useExternal = false

		ta := textarea.New()
		ta.SetValue(m.original)
		ta.Focus()
		m.textarea = ta

		return m, textarea.Blink
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-4)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4
		}
	}

	if !m.useExternal {
		var taCmd tea.Cmd
		m.textarea, taCmd = m.textarea.Update(msg)
		cmds = append(cmds, taCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *YAMLEditor) View() string {
	header := m.headerView()
	footer := m.footerView()

	var content string
	if m.useExternal {
		editor := utils.GetPreferredEditor()
		content = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(m.styles.BorderColor)).
			Padding(1).
			BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render(fmt.Sprintf("Opening %s...\n\nFile: %s\n\nPress 'esc' to cancel if editor doesn't launch", editor, m.tempFile))
	} else {
		content = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(m.styles.BorderColor)).
			Padding(1).
			BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
			Background(lipgloss.Color(customstyles.BackgroundColor)).
			Render(m.textarea.View())
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

func (m *YAMLEditor) headerView() string {
	title := m.styles.TitleBar.Render(m.title + " [EDITING]")
	return title
}

func (m *YAMLEditor) footerView() string {
	var helpText string

	if m.useExternal {
		helpText = "Esc: Cancel"
		if m.customHelp != "" {
			helpText = m.customHelp
		}
	} else {
		helpText = "Ctrl+S: Save • Esc: Cancel"
		if m.customHelp != "" {
			helpText = m.customHelp
		}

		isModified := m.textarea.Value() != m.original
		if isModified {
			helpText += " • [MODIFIED]"
		}
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.HelpTextColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(helpText)

	return help
}

func (m *YAMLEditor) GetContent() string {
	if m.useExternal {
		content, err := m.readEditedContent()
		if err != nil {
			return m.original
		}
		return content
	}
	return m.textarea.Value()
}

func (m *YAMLEditor) IsModified() bool {
	return m.GetContent() != m.original
}

func (m *YAMLEditor) UseExternal() bool {
	return m.useExternal
}

func (m *YAMLEditor) GetTempFile() string {
	return m.tempFile
}

type SaveMsg struct {
	Content    string
	Title      string
	Original   string
	IsModified bool
}

type CancelMsg struct{}

func createTempFile(content string) (string, error) {
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("k8s-tui-edit-%d.yaml", os.Getpid()))

	err := os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		return "", err
	}

	return tempFile, nil
}

func (m *YAMLEditor) launchExternalEditor() tea.Cmd {
	editor := utils.GetPreferredEditor()

	cmd := exec.Command(editor, m.tempFile)
	m.editorCmd = cmd

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		if err != nil {
			return ExternalEditorErrorMsg{Error: err}
		}

		if cmd.ProcessState != nil {
			exitCode := cmd.ProcessState.ExitCode()

			switch editor {
			case "vim", "nvim":
				if exitCode != 0 {
					return ExternalEditorErrorMsg{Error: fmt.Errorf("editor exited with code %d (likely cancelled)", exitCode)}
				}
			case "nano":
				if exitCode == 1 {
					return ExternalEditorErrorMsg{Error: fmt.Errorf("editor cancelled")}
				}
			}
		}

		return ExternalEditorDoneMsg{}
	})
}

func (m *YAMLEditor) readEditedContent() (string, error) {
	if m.tempFile == "" {
		return "", fmt.Errorf("no temporary file")
	}

	content, err := os.ReadFile(m.tempFile)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func (m *YAMLEditor) cleanupTempFile() {
	if m.tempFile != "" {
		os.Remove(m.tempFile)
		m.tempFile = ""
	}
}

type ExternalEditorDoneMsg struct{}
type ExternalEditorErrorMsg struct {
	Error error
}
