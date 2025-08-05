package global

import (
	"fmt"
	"os"
	"path/filepath"
)

var Margin = 2

var ScreenWidth = 0
var ScreenHeight = 0

var HeaderSize = 0

var kubeconfigsDefaultLocation = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Erro ao obter o diretório home:", err)
		return ""
	}
	return filepath.Join(home, ".kube")
}()

func GetKubeconfigsLocations() []string{
  KubeconfigsLocations := []string{kubeconfigsDefaultLocation}
  return KubeconfigsLocations
}

