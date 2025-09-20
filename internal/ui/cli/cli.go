package cli

import (
	"flag"
	"github.com/otavioCosta2110/k8s-tui/internal/config"
)

type Config struct {
	KubeconfigPath string
	Namespace      string
	PluginDir      string
}

func ParseFlags() Config {
	var cfg Config

	// Load app config to get default plugin directory
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
