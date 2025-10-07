package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/otavioCosta2110/k8s-tui/internal/app/config"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
)

type Config struct {
	KubeconfigPath string
	Namespace      string
	PluginDir      string
	PluginArgs     map[string]string
}

func ParseFlags() Config {
	var cfg Config

	appConfig, err := config.LoadAppConfig()
	defaultPluginDir := "./plugins"
	if err == nil {
		defaultPluginDir = appConfig.PluginDir
	}

	// Parse all command line arguments manually to handle both known and plugin flags
	args := os.Args[1:] // Skip program name
	cfg.PluginArgs = make(map[string]string)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			flagName := strings.TrimPrefix(arg, "--")
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				flagValue := args[i+1]
				// Check if it's a known flag
				switch flagName {
				case "kubeconfig":
					cfg.KubeconfigPath = flagValue
				case "namespace":
					cfg.Namespace = flagValue
				case "plugin-dir":
					cfg.PluginDir = flagValue
				default:
					// Treat as plugin argument
					cfg.PluginArgs[flagName] = flagValue
				}
				i++ // Skip the value
			} else {
				// Flag without value
				switch flagName {
				case "help", "h":
					// Handle help flag
				default:
					cfg.PluginArgs[flagName] = "true"
				}
			}
		}
	}

	// Set defaults if not provided
	if cfg.PluginDir == "" {
		cfg.PluginDir = defaultPluginDir
	}

	return cfg
}

func HandlePluginArgs(pluginManager *plugins.PluginManager, pluginArgs map[string]string) error {
	fmt.Printf("DEBUG: HandlePluginArgs called with args: %v\n", pluginArgs)
	if pluginManager == nil {
		return nil
	}

	api := pluginManager.GetAPI()
	if api == nil {
		return nil
	}

	for argName, argValue := range pluginArgs {
		fmt.Printf("DEBUG: Checking arg %s\n", argName)
		if api.HasCLIArgument(argName) {
			fmt.Printf("DEBUG: Executing arg %s with value %s\n", argName, argValue)
			if err := api.ExecuteCLIArgument(argName, argValue); err != nil {
				fmt.Printf("DEBUG: Error executing arg: %v\n", err)
				return err
			}
			fmt.Printf("DEBUG: Arg executed successfully\n")
		} else {
			fmt.Printf("DEBUG: Arg %s not registered\n", argName)
		}
	}

	return nil
}
