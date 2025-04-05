package kubernetes

import (
	"os"
	"path/filepath"
)

func GetKubeconfigsLocations() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{"."} 
	}

	return []string{
		filepath.Join(home, ".kube"),
	}
}
