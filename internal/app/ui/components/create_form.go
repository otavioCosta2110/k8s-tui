package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
)

type Field struct {
	Name         string
	Placeholder  string
	DefaultValue string
	Row          int
}

type CreateForm struct {
	inputs   []textinput.Model
	fields   []Field
	current  int
	title    string
	onSubmit func(map[string]string) tea.Msg
	onCancel func() tea.Msg
}

type CreateSubmitMsg struct {
	ResourceType string
	Values       map[string]string
}

type CreateCancelMsg struct{}

func NewCreateForm(title string, resourceType string, fields []Field) *CreateForm {
	inputs := make([]textinput.Model, len(fields))

	for i, field := range fields {
		ti := textinput.New()
		ti.Placeholder = field.Placeholder
		ti.CharLimit = 100
		ti.Width = 50
		ti.TextStyle = lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Foreground(lipgloss.Color(customstyles.TextColor))
		ti.PlaceholderStyle = lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Foreground(lipgloss.Color(customstyles.HelpTextColor))
		ti.Cursor.TextStyle = lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).Foreground(lipgloss.Color(customstyles.TextColor))

		if field.DefaultValue != "" {
			ti.SetValue(field.DefaultValue)
		}
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}

	onSubmit := func(values map[string]string) tea.Msg {
		return CreateSubmitMsg{
			ResourceType: resourceType,
			Values:       values,
		}
	}
	onCancel := func() tea.Msg {
		return CreateCancelMsg{}
	}

	return &CreateForm{
		inputs:   inputs,
		fields:   fields,
		current:  0,
		title:    title,
		onSubmit: onSubmit,
		onCancel: onCancel,
	}
}

func (f *CreateForm) Init() tea.Cmd {
	return textinput.Blink
}

func (f *CreateForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if f.current < len(f.inputs)-1 {
				f.inputs[f.current].Blur()
				f.current++
				f.inputs[f.current].Focus()
				return f, textinput.Blink
			} else {
				values := make(map[string]string)
				for i, input := range f.inputs {
					values[f.fields[i].Name] = input.Value()
				}
				if f.onSubmit != nil {
					return f, func() tea.Msg {
						return f.onSubmit(values)
					}
				}
			}
		case "esc":
			if f.onCancel != nil {
				return f, func() tea.Msg {
					return f.onCancel()
				}
			}
		case "tab", "shift+tab":
			f.inputs[f.current].Blur()
			if msg.String() == "shift+tab" {
				f.current--
				if f.current < 0 {
					f.current = len(f.inputs) - 1
				}
			} else {
				f.current++
				if f.current >= len(f.inputs) {
					f.current = 0
				}
			}
			f.inputs[f.current].Focus()
			return f, textinput.Blink
		default:
			f.inputs[f.current], cmd = f.inputs[f.current].Update(msg)
		}
	}

	return f, cmd
}

func (f *CreateForm) View() string {
	W := styles.ScreenWidth + styles.Margin
	H := styles.ScreenHeight + styles.HeaderSize + styles.TabBarSize + 1

	basePadding := H/2 - 8
	fieldAdjustment := (len(f.fields) - 3) * 2
	topPadding := max(basePadding-fieldAdjustment, 1)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(customstyles.AccentColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Width(W/2 + styles.Margin).
		PaddingTop(topPadding).
		Render(f.title)

	rowMap := make(map[int][]int)
	for i, field := range f.fields {
		rowMap[field.Row] = append(rowMap[field.Row], i)
	}

	for _, indices := range rowMap {
		count := len(indices)
		for _, idx := range indices {
			switch count {
			case 1:
				f.inputs[idx].Width = W/2 - 6
			case 2:
				f.inputs[idx].Width = W/4 - 4
			}
		}
	}

	var inputsView string
	for row := 0; ; row++ {
		indices, exists := rowMap[row]
		if !exists {
			break
		}
		count := len(indices)
		switch count {
		case 1:
			idx := indices[0]
			label := lipgloss.NewStyle().
				Foreground(lipgloss.Color(customstyles.TextColor)).
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				Width(W/2 + 2).
				Align(lipgloss.Left).
				Render(f.fields[idx].Placeholder + ":")

			inputView := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Background(lipgloss.Color(customstyles.BackgroundColor)).
				BorderForeground(lipgloss.Color(customstyles.BorderColor)).
				BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
				Width(W / 2).
				Render(f.inputs[idx].View())

			field := lipgloss.JoinVertical(lipgloss.Left, label, inputView)
			if inputsView != "" {
				inputsView = lipgloss.JoinVertical(lipgloss.Left, inputsView, "", field)
			} else {
				inputsView = field
			}
		case 2:
			var rowView string
			for _, idx := range indices {
				label := lipgloss.NewStyle().
					Foreground(lipgloss.Color(customstyles.TextColor)).
					Background(lipgloss.Color(customstyles.BackgroundColor)).
					Width(W / 4).
					Align(lipgloss.Left).
					Render(f.fields[idx].Placeholder + ":")

				inputView := lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					Background(lipgloss.Color(customstyles.BackgroundColor)).
					BorderForeground(lipgloss.Color(customstyles.BorderColor)).
					BorderBackground(lipgloss.Color(customstyles.BackgroundColor)).
					Width(W/4 - 2).
					Render(f.inputs[idx].View())

				field := lipgloss.JoinVertical(lipgloss.Left, label, inputView)
				if rowView != "" {
					rowViewStyled := lipgloss.NewStyle().Background(lipgloss.Color(customstyles.BackgroundColor)).PaddingRight(2).Render(rowView)
					rowView = lipgloss.JoinHorizontal(lipgloss.Top, rowViewStyled, field)
				} else {
					rowView = field
				}
			}
			if inputsView != "" {
				inputsView = lipgloss.JoinVertical(lipgloss.Left, inputsView, "", rowView)
			} else {
				inputsView = rowView
			}
		}
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HelpTextColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		PaddingTop(2).
		Width(W/2 + styles.Margin).
		Render("Tab/Shift+Tab: Navigate • Enter: Next/Submit • Esc: Cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		inputsView,
		help,
	)

	return lipgloss.NewStyle().
		Width(W).
		Height(H).
		Align(lipgloss.Center).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Render(content)
}
