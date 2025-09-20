package cli

import (
	"flag"
	"github.com/otavioCosta2110/k8s-tui/pkg/config"
)

type Config struct {
	KubeconfigPath string
	Namespace      string
	PluginDir      string
}

func ParseFlags() Config {
	var cfg Config

	appConfig, err := config.LoadAppConfig()
	defaultPluginDir := "./plugins"
	if err == nil {
		defaultPluginDir = appConfig.PluginDir
	}

	flag.StringVar(&cfg.KubeconfigPath, "kubeconfig", "", "path to the kubeconfig file")
	flag.StringVar(&cfg.Namespace, "namespace", "", "namespace to use")
	flag.StringVar(&cfg.PluginDir, "plugin-dir", defaultPluginDir, "directory containing plugin files")

	flag.Parse()

	return cfg
}
