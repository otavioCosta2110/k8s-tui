package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadColorScheme(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".config", "k8s-tui")
	os.MkdirAll(configDir, 0755)

	testScheme := `{
  "border_color": "#ff0000",
  "accent_color": "#00ff00",
  "header_color": "#0000ff",
  "error_color": "#ffff00",
  "selection_background": "#ff00ff",
  "selection_foreground": "#ffffff",
  "background_color": "#123456",
  "text_color": "#abcdef"
}`

	configFile := filepath.Join(configDir, "colorscheme.json")
	err := os.WriteFile(configFile, []byte(testScheme), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	scheme, err := LoadColorScheme()
	if err != nil {
		t.Fatalf("Failed to load color scheme: %v", err)
	}

	if scheme.BorderColor != "#ff0000" {
		t.Errorf("Expected BorderColor #ff0000, got %s", scheme.BorderColor)
	}
	if scheme.BackgroundColor != "#123456" {
		t.Errorf("Expected BackgroundColor #123456, got %s", scheme.BackgroundColor)
	}
	if scheme.TextColor != "#abcdef" {
		t.Errorf("Expected TextColor #abcdef, got %s", scheme.TextColor)
	}
}

func TestDefaultColorScheme(t *testing.T) {
	scheme := DefaultColorScheme()

	if scheme.BorderColor == "" {
		t.Error("Default BorderColor should not be empty")
	}
	if scheme.BackgroundColor != "#000000" {
		t.Errorf("Default BackgroundColor should be #000000, got %s", scheme.BackgroundColor)
	}
}
