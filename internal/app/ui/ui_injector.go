package ui

import (
	"strings"

	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
)

type UIInjector struct {
	injections map[string][]plugins.UIInjectionPoint
}

func NewUIInjector() *UIInjector {
	return &UIInjector{
		injections: make(map[string][]plugins.UIInjectionPoint),
	}
}

func (ui *UIInjector) AddInjection(location string, injection plugins.UIInjectionPoint) {
	ui.injections[location] = append(ui.injections[location], injection)
}

func (ui *UIInjector) GetInjections(location string) []plugins.UIInjectionPoint {
	return ui.injections[location]
}

func (ui *UIInjector) RenderInjections(location string) string {
	injections := ui.GetInjections(location)
	if len(injections) == 0 {
		return ""
	}

	var rendered []string
	for _, injection := range injections {
		switch injection.Component.Type {
		case "text":
			if content, ok := injection.Component.Config["content"].(string); ok {
				rendered = append(rendered, content)
			}
		default:
			if content, ok := injection.Component.Config["content"].(string); ok {
				rendered = append(rendered, content)
			}
		}
	}

	return strings.Join(rendered, " | ")
}
