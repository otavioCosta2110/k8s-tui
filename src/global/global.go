package global

import (
	"fmt"
	"os"
	"path/filepath"
)

var Colors = struct {
	Blue, Pink string
}{
	Blue: "#00b8ff",
  Pink: "#f29bdc",
}

var Margin = 2

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
