package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles"
	customstyles "github.com/otavioCosta2110/k8s-tui/internal/app/ui/styles/custom_styles"
)

type Event struct {
	Type      string
	Reason    string
	Age       string
	From      string
	Message   string
	Object    string
	Namespace string
	Count     int32
	FirstSeen string
	LastSeen  string
}

type EventsModel struct {
	viewport viewport.Model
	events   []Event
	title    string
	loading  bool
	error    error
	width    int
	height   int
}

type EventsLoadedMsg struct {
	Events []Event
	Error  error
}

func NewEventsModel(title string) *EventsModel {
	vp := viewport.New(styles.ScreenWidth, styles.ScreenHeight-4)
	vp.Style = lipgloss.NewStyle().
		Background(lipgloss.Color(customstyles.BackgroundColor))

	return &EventsModel{
		viewport: vp,
		title:    title,
		loading:  true,
		events:   make([]Event, 0),
		width:    styles.ScreenWidth,
		height:   styles.ScreenHeight - 4,
	}
}

func (m *EventsModel) SetEvents(events []Event) {
	m.events = events
	m.loading = false
	m.updateContent()
}

func (m *EventsModel) SetError(err error) {
	m.error = err
	m.loading = false
	m.updateContent()
}

func (m *EventsModel) SetLoading(loading bool) {
	m.loading = loading
	if loading {
		m.error = nil
	}
	m.updateContent()
}

func (m *EventsModel) Init() tea.Cmd {
	return nil
}

func (m *EventsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 4
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 4
		m.updateContent()
	case EventsLoadedMsg:
		if msg.Error != nil {
			m.SetError(msg.Error)
		} else {
			m.SetEvents(msg.Events)
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *EventsModel) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.error != nil {
		return m.renderError()
	}

	if len(m.events) == 0 {
		return m.renderEmpty()
	}

	return m.viewport.View()
}

func (m *EventsModel) updateContent() {
	if m.loading {
		m.viewport.SetContent(m.renderLoading())
		return
	}

	if m.error != nil {
		m.viewport.SetContent(m.renderError())
		return
	}

	if len(m.events) == 0 {
		m.viewport.SetContent(m.renderEmpty())
		return
	}

	m.viewport.SetContent(m.renderEvents())
}

func (m *EventsModel) renderLoading() string {
	loadingText := "Loading events..."
	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HeaderLoadingColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Bold(true)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		loadingStyle.Render(loadingText),
	)
}

func (m *EventsModel) renderError() string {
	errorText := fmt.Sprintf("Error loading events: %v", m.error)
	errorStyle := customstyles.ErrorStyle().
		Background(lipgloss.Color(customstyles.BackgroundColor))

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		errorStyle.Render(errorText),
	)
}

func (m *EventsModel) renderEmpty() string {
	emptyText := "No events found for this resource."
	emptyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.HelpTextColor)).
		Background(lipgloss.Color(customstyles.BackgroundColor)).
		Align(lipgloss.Center).
		Width(styles.ScreenWidth).
		Italic(true)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		emptyStyle.Render(emptyText),
	)
}

func (m *EventsModel) renderEvents() string {
	var builder strings.Builder

	// Title
	titleStyle := customstyles.TitleStyle()
	builder.WriteString(titleStyle.Render(m.title))
	builder.WriteString("\n\n")

	// Calculate column widths based on content
	maxTypeWidth := 6
	maxReasonWidth := 10
	maxAgeWidth := 8
	maxFromWidth := 10
	maxObjectWidth := 15

	for _, event := range m.events {
		if len(event.Type) > maxTypeWidth {
			maxTypeWidth = len(event.Type)
		}
		if len(event.Reason) > maxReasonWidth {
			maxReasonWidth = len(event.Reason)
		}
		if len(event.Age) > maxAgeWidth {
			maxAgeWidth = len(event.Age)
		}
		if len(event.From) > maxFromWidth {
			maxFromWidth = len(event.From)
		}
		if len(event.Object) > maxObjectWidth {
			maxObjectWidth = len(event.Object)
		}
	}

	// Limit column widths
	maxTypeWidth = min(maxTypeWidth, 10)
	maxReasonWidth = min(maxReasonWidth, 15)
	maxAgeWidth = min(maxAgeWidth, 12)
	maxFromWidth = min(maxFromWidth, 15)
	maxObjectWidth = min(maxObjectWidth, 25)

	// Calculate message width (remaining space)
	messageWidth := m.width - maxTypeWidth - maxReasonWidth - maxAgeWidth - maxFromWidth - maxObjectWidth - 5 // 10 for spacing and borders
	messageWidth = max(messageWidth, 20)

	// Table headers
	headers := []string{
		"TYPE",
		"REASON",
		"AGE",
		"FROM",
		"MESSAGE",
		"OBJECT",
	}

	widths := []int{maxTypeWidth, maxReasonWidth, maxAgeWidth, maxFromWidth, messageWidth, maxObjectWidth}

	whitespaceStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(customstyles.BackgroundColor))

	headerRow := ""
	for i, header := range headers {
		if i > 0 {
			headerRow += whitespaceStyle.Render(" ")
		}
		style := lipgloss.NewStyle().
			Width(widths[i]).
			MaxWidth(widths[i]).
			Bold(true).
			Foreground(lipgloss.Color(customstyles.AccentColor)).
			Background(lipgloss.Color(customstyles.BackgroundColor))
		headerRow += style.Render(header)
	}
	builder.WriteString(headerRow)
	builder.WriteString("\n")

	// Separator
	separator := strings.Repeat("─", m.width)
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(customstyles.BorderColor))
	builder.WriteString(separatorStyle.Render(separator))
	builder.WriteString("\n")

	// Event rows
	for _, event := range m.events {
		row := ""

		// Truncate long fields
		typeVal := event.Type
		if len(typeVal) > maxTypeWidth {
			typeVal = typeVal[:maxTypeWidth-3] + "..."
		}

		reason := event.Reason
		if len(reason) > maxReasonWidth {
			reason = reason[:maxReasonWidth-3] + "..."
		}

		age := event.Age
		if len(age) > maxAgeWidth {
			age = age[:maxAgeWidth-3] + "..."
		}

		from := event.From
		if len(from) > maxFromWidth {
			from = from[:maxFromWidth-3] + "..."
		}

		message := event.Message
		if len(message) > messageWidth {
			message = message[:messageWidth-3] + "..."
		}

		object := event.Object
		if len(object) > maxObjectWidth {
			object = object[:maxObjectWidth-3] + "..."
		}

		values := []string{typeVal, reason, age, from, message, object}

		for i, value := range values {
			if i > 0 {
				row += whitespaceStyle.Render(" ")
			}
			style := lipgloss.NewStyle().
				Width(widths[i]).
				MaxWidth(widths[i]).
				Background(lipgloss.Color(customstyles.BackgroundColor))

			// Color code event types
			switch event.Type {
			case "Warning":
				row += customstyles.ErrorStyle().Render(style.Render(value))
			case "Normal":
				row += style.Foreground(lipgloss.Color("10")).Render(value) // Green
			default:
				row += style.Render(value)
			}
		}

		builder.WriteString(row)
		builder.WriteString("\n")
	}

	return builder.String()
}

func (m *EventsModel) GotoTop() {
	m.viewport.GotoTop()
}

func (m *EventsModel) GotoBottom() {
	m.viewport.GotoBottom()
}

func (m *EventsModel) LineDown(n int) {
	m.viewport.ScrollDown(n)
}

func (m *EventsModel) LineUp(n int) {
	m.viewport.ScrollUp(n)
}

func (m *EventsModel) ViewportHeight() int {
	return m.viewport.Height
}

func (m *EventsModel) ViewportWidth() int {
	return m.viewport.Width
}
