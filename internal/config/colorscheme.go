package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ColorScheme struct {
	BorderColor         string `json:"border_color"`
	AccentColor         string `json:"accent_color"`
	HeaderColor         string `json:"header_color"`
	ErrorColor          string `json:"error_color"`
	SelectionBackground string `json:"selection_background"`
	SelectionForeground string `json:"selection_foreground"`
	TextColor           string `json:"text_color,omitempty"`
	BackgroundColor     string `json:"background_color,omitempty"`
	YAMLKeyColor        string `json:"yaml_key_color,omitempty"`
	YAMLValueColor      string `json:"yaml_value_color,omitempty"`
	YAMLTitleColor      string `json:"yaml_title_color,omitempty"`
	HelpTextColor       string `json:"help_text_color,omitempty"`
}

func DefaultColorScheme() ColorScheme {
	return ColorScheme{
		BorderColor:         "#00b8ff",
		AccentColor:         "#f29bdc",
		HeaderColor:         "#7D56F4",
		ErrorColor:          "#FF0000",
		SelectionBackground: "#00b8ff",
		SelectionForeground: "#000000",
		BackgroundColor:     "#000000",
		YAMLKeyColor:        "#5E9AFF",
		YAMLValueColor:      "",
		YAMLTitleColor:      "#FAFAFA",
		HelpTextColor:       "#757575",
	}
}

func LoadColorScheme() (ColorScheme, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "k8s-tui")
	configFile := filepath.Join(configDir, "colorscheme.json")

	scheme := DefaultColorScheme()

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return scheme, err
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultScheme := DefaultColorScheme()
		data, err := json.MarshalIndent(defaultScheme, "", "  ")
		if err != nil {
			return scheme, err
		}
		if err := os.WriteFile(configFile, data, 0644); err != nil {
			return scheme, err
		}
		return scheme, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return scheme, err
	}

	var loadedScheme ColorScheme
	if err := json.Unmarshal(data, &loadedScheme); err != nil {
		return scheme, err
	}

	if loadedScheme.BorderColor != "" {
		scheme.BorderColor = loadedScheme.BorderColor
	}
	if loadedScheme.AccentColor != "" {
		scheme.AccentColor = loadedScheme.AccentColor
	}
	if loadedScheme.HeaderColor != "" {
		scheme.HeaderColor = loadedScheme.HeaderColor
	}
	if loadedScheme.ErrorColor != "" {
		scheme.ErrorColor = loadedScheme.ErrorColor
	}
	if loadedScheme.SelectionBackground != "" {
		scheme.SelectionBackground = loadedScheme.SelectionBackground
	}
	if loadedScheme.SelectionForeground != "" {
		scheme.SelectionForeground = loadedScheme.SelectionForeground
	}
	if loadedScheme.TextColor != "" {
		scheme.TextColor = loadedScheme.TextColor
	}
	if loadedScheme.BackgroundColor != "" {
		scheme.BackgroundColor = loadedScheme.BackgroundColor
	}
	if loadedScheme.YAMLKeyColor != "" {
		scheme.YAMLKeyColor = loadedScheme.YAMLKeyColor
	}
	if loadedScheme.YAMLValueColor != "" {
		scheme.YAMLValueColor = loadedScheme.YAMLValueColor
	}
	if loadedScheme.YAMLTitleColor != "" {
		scheme.YAMLTitleColor = loadedScheme.YAMLTitleColor
	}
	if loadedScheme.HelpTextColor != "" {
		scheme.HelpTextColor = loadedScheme.HelpTextColor
	}

	return scheme, nil
}
