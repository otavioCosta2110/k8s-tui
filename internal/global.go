package global

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/otavioCosta2110/k8s-tui/internal/logger"
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
