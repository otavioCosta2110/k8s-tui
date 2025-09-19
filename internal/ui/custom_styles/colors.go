package customstyles

import (
	"otaviocosta2110/k8s-tui/internal/config"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	BorderColor string

	AccentColor string

	HeaderColor string

	ErrorColor string

	SelectionBackground string

	SelectionForeground string

	TextColor lipgloss.Color

	BackgroundColor string

	YAMLKeyColor string

	YAMLValueColor string

	YAMLTitleColor string

	HelpTextColor string

	HeaderValueColor string

	HeaderLoadingColor string

	ResourceIcons map[string]string
)

func InitColors() error {
	scheme, err := config.LoadColorScheme()
	if err != nil {
		return err
	}

	BorderColor = scheme.BorderColor
	AccentColor = scheme.AccentColor
	HeaderColor = scheme.HeaderColor
	ErrorColor = scheme.ErrorColor
	SelectionBackground = scheme.SelectionBackground
	SelectionForeground = scheme.SelectionForeground

	if scheme.TextColor != "" {
		TextColor = lipgloss.Color(scheme.TextColor)
	} else {
		TextColor = lipgloss.Color(termenv.ForegroundColor().Sequence(true))
	}

	if scheme.BackgroundColor != "" {
		BackgroundColor = scheme.BackgroundColor
	} else {
		BackgroundColor = "#000000"
	}

	YAMLKeyColor = scheme.YAMLKeyColor
	YAMLValueColor = scheme.YAMLValueColor
	YAMLTitleColor = scheme.YAMLTitleColor
	HelpTextColor = scheme.HelpTextColor
	HeaderValueColor = scheme.HeaderValueColor
	HeaderLoadingColor = scheme.HeaderLoadingColor

	ResourceIcons = map[string]string{
		"Pods":                   "󰀵",
		"Deployments":            "󰜴",
		"Services":               "󰖟",
		"Ingresses":              "󰜏",
		"ConfigMaps":             "󰈙",
		"Secrets":                "󰌿",
		"ReplicaSets":            "󰑖",
		"Jobs":                   "󰜎",
		"CronJobs":               "󰥔",
		"DaemonSets":             "󰜙",
		"StatefulSets":           "󰋊",
		"Nodes":                  "󰒍",
		"Namespaces":             "󰉋",
		"PersistentVolumes":      "󰋊",
		"PersistentVolumeClaims": "󰋊",
		"ServiceAccounts":        "󰀄",
		"ResourceList":           "󰒋",
		"Workloads":              "󰜄",
		"Networking":             "󰖟",
		"Configuration":          "󰒓",
		"Infrastructure":         "󰒍",
		"Navigation":             "󰍉",
	}

	return nil
}
