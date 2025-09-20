package global

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
)

var Margin = 2

var ScreenWidth = 0
var ScreenHeight = 0

var HeaderSize = 0
var IsHeaderActive = false

var TabBarSize = 3
var IsTabBarActive = false

var kubeconfigsDefaultLocation = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Error(fmt.Sprintf("Erro ao obter o diretório home: %v", err))
		return ""
	}
	return filepath.Join(home, ".kube")
}()

func GetKubeconfigsLocations() []string {
	KubeconfigsLocations := []string{kubeconfigsDefaultLocation}
	return KubeconfigsLocations
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
