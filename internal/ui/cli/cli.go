package cli

import (
	"flag"
)

type Config struct {
	KubeconfigPath string
	Namespace      string
}

func ParseFlags() Config {
	var cfg Config

	flag.StringVar(&cfg.KubeconfigPath, "kubeconfig", "", "path to the kubeconfig file")
	flag.StringVar(&cfg.Namespace, "namespace", "", "namespace to use")

	flag.Parse()

	return cfg
}
