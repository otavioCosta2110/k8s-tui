package utils

import (
	"fmt"
	"os"
	"os/exec"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
)

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
