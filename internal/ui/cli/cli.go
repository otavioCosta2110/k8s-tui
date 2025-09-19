package cli

import (
	"flag"
)

type Config struct {
	KubeconfigPath string
	Namespace      string
	PluginDir      string
}

func ParseFlags() Config {
	var cfg Config

	flag.StringVar(&cfg.KubeconfigPath, "kubeconfig", "", "path to the kubeconfig file")
	flag.StringVar(&cfg.Namespace, "namespace", "", "namespace to use")
	flag.StringVar(&cfg.PluginDir, "plugin-dir", "./plugins", "directory containing plugin files")

	flag.Parse()

	return cfg
}
