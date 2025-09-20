package utils

import (
	"os"
	"os/exec"
	"strings"
)

func WriteString(filePath string, data string) error {
	return os.WriteFile(filePath, []byte(data), 0644)
}

func WriteStringArray(filePath string, data []string) error {
	content := strings.Join(data, "\n")
	return os.WriteFile(filePath, []byte(content), 0644)
}

func WriteStringNewLine(filePath string, data string) error {
	prevContent, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	prevContentStr := string(prevContent)
	content := strings.TrimSpace(prevContentStr) + "\n" + strings.TrimSpace(data) + "\n"
	return os.WriteFile(filePath, []byte(content), 0644)
}

func GetPreferredEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}

	commonEditors := []string{"vim", "nvim", "nano", "emacs", "code", "subl"}
	for _, editor := range commonEditors {
		if _, err := exec.LookPath(editor); err == nil {
			return editor
		}
	}

	return "vi"
}
