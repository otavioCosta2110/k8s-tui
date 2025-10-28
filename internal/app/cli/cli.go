package cli

import (
	"os"
	"strings"

	"github.com/otavioCosta2110/k8s-tui/internal/app/config"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
)

type Config struct {
	KubeconfigPaths []string
	Namespace       string
	PluginDir       string
	PluginArgs      map[string]string
}

func ParseFlags() Config {
	var cfg Config

	appConfig, err := config.LoadAppConfig()
	defaultPluginDir := "./plugins"
	if err == nil {
		defaultPluginDir = appConfig.PluginDir
	}

	args := os.Args[1:]
	cfg.PluginArgs = make(map[string]string)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if after, ok := strings.CutPrefix(arg, "--"); ok {
			flagName := after
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				flagValue := args[i+1]
				switch flagName {
				case "kubeconfig":
					cfg.KubeconfigPaths = append(cfg.KubeconfigPaths, flagValue)
				case "namespace":
					cfg.Namespace = flagValue
				case "plugin-dir":
					cfg.PluginDir = flagValue
				default:
					cfg.PluginArgs[flagName] = flagValue
				}
				i++
			} else {
				switch flagName {
				case "help", "h":
				default:
					cfg.PluginArgs[flagName] = "true"
				}
			}
		}
	}

	if cfg.PluginDir == "" {
		cfg.PluginDir = defaultPluginDir
	}

	return cfg
}

func HandlePluginArgs(pluginManager *plugins.PluginManager, pluginArgs map[string]string) error {
	if pluginManager == nil {
		return nil
	}

	api := pluginManager.GetAPI()
	if api == nil {
		return nil
	}

	for argName, argValue := range pluginArgs {
		if api.HasCLIArgument(argName) {
			if err := api.ExecuteCLIArgument(argName, argValue); err != nil {
				return err
			}
		}
	}

	return nil
}
