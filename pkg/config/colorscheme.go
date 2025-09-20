package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	HeaderValueColor    string `json:"header_value_color,omitempty"`
	HeaderLoadingColor  string `json:"header_loading_color,omitempty"`
}

type AppConfig struct {
	Theme            string            `json:"theme"`
	RefreshInterval  int               `json:"refresh_interval_seconds"`
	AutoRefresh      bool              `json:"auto_refresh"`
	DefaultNamespace string            `json:"default_namespace"`
	PluginDir        string            `json:"plugin_dir,omitempty"`
	KeyBindings      map[string]string `json:"key_bindings,omitempty"`
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
		HeaderValueColor:    "#A1EFD3",
		HeaderLoadingColor:  "#FFA500",
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
	if loadedScheme.HeaderValueColor != "" {
		scheme.HeaderValueColor = loadedScheme.HeaderValueColor
	}
	if loadedScheme.HeaderLoadingColor != "" {
		scheme.HeaderLoadingColor = loadedScheme.HeaderLoadingColor
	}

	return scheme, nil
}

func DefaultAppConfig() AppConfig {
	return AppConfig{
		Theme:            "default",
		RefreshInterval:  10,
		AutoRefresh:      true,
		DefaultNamespace: "default",
		PluginDir:        "~/.local/share/k8s-tui/plugins",
		KeyBindings: map[string]string{
			"quit":      "q",
			"help":      "?",
			"refresh":   "r",
			"back":      "[",
			"forward":   "]",
			"new_tab":   "ctrl+t",
			"close_tab": "ctrl+w",
			"quick_nav": "g",
		},
	}
}

func LoadAppConfig() (AppConfig, error) {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "k8s-tui")
	configFile := filepath.Join(configDir, "config.json")

	config := DefaultAppConfig()

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return config, err
	}

	if err := setupThemes(); err != nil {
		return config, err
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return config, err
		}
		if err := os.WriteFile(configFile, data, 0644); err != nil {
			return config, err
		}
		return config, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return config, err
	}

	var loadedConfig AppConfig
	if err := json.Unmarshal(data, &loadedConfig); err != nil {
		return config, err
	}

	if loadedConfig.RefreshInterval == 0 {
		loadedConfig.RefreshInterval = config.RefreshInterval
	}
	if loadedConfig.KeyBindings == nil {
		loadedConfig.KeyBindings = config.KeyBindings
	}
	if loadedConfig.PluginDir == "" {
		loadedConfig.PluginDir = config.PluginDir
	}

	// Expand ~ to home directory
	if strings.HasPrefix(loadedConfig.PluginDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			loadedConfig.PluginDir = filepath.Join(homeDir, loadedConfig.PluginDir[2:])
		}
	}

	return loadedConfig, nil
}

func setupThemes() error {
	themesDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "k8s-tui", "themes")

	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return err
	}

	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execDir := filepath.Dir(execPath)

	sourceThemesDir := filepath.Join(execDir, "colorschemes")

	if _, err := os.Stat(sourceThemesDir); os.IsNotExist(err) {
		if cwd, err := os.Getwd(); err == nil {
			sourceThemesDir = filepath.Join(cwd, "colorschemes")
		}
	}

	if _, err := os.Stat(sourceThemesDir); err == nil {
		files, err := os.ReadDir(sourceThemesDir)
		if err != nil {
			return err
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".json") && file.Name() != "colorscheme.json" {
				srcPath := filepath.Join(sourceThemesDir, file.Name())
				dstPath := filepath.Join(themesDir, file.Name())

				if err := copyFile(srcPath, dstPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func LoadTheme(themeName string) (ColorScheme, error) {
	themesDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "k8s-tui", "themes")
	themeFile := filepath.Join(themesDir, themeName+".json")

	if _, err := os.Stat(themeFile); os.IsNotExist(err) {
		return DefaultColorScheme(), nil
	}

	data, err := os.ReadFile(themeFile)
	if err != nil {
		return DefaultColorScheme(), err
	}

	var scheme ColorScheme
	if err := json.Unmarshal(data, &scheme); err != nil {
		return DefaultColorScheme(), err
	}

	return scheme, nil
}
