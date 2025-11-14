package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/otavioCosta2110/k8s-tui/internal/app/config"
	"github.com/otavioCosta2110/k8s-tui/pkg/plugins"
)

var (
	PluginArgs          map[string]string
	KubeconfigPaths     []string
	Namespace           string
	PluginDir           string
	SessionFile         string
	SessionClusters     []SessionCluster
	pluginArgsProcessed = false
)

type Config struct {
	KubeconfigPaths []string
	Namespace       string
	PluginDir       string
	PluginArgs      map[string]string
	SessionFile     string
	SessionClusters []SessionCluster
}

// SessionCluster represents a cluster in a session file
type SessionCluster struct {
	Index      int    `json:"Index"`
	Name       string `json:"Name"`
	Namespace  string `json:"Namespace"`
	Kubeconfig string `json:"Kubeconfig"`
	IsActive   bool   `json:"IsActive"`
}

// SessionData represents the structure of a session file
type SessionData struct {
	Clusters       []SessionCluster `json:"clusters"`
	CurrentCluster SessionCluster   `json:"current_cluster"`
}

// parseSessionFile parses a session file and returns cluster configurations
func parseSessionFile(sessionPath string) ([]SessionCluster, error) {
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %v", err)
	}

	var sessionData SessionData
	if err := json.Unmarshal(data, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to parse session JSON: %v", err)
	}

	return sessionData.Clusters, nil
}

func ParseFlags() Config {
	var cfg Config

	appConfig, err := config.LoadAppConfig()
	defaultPluginDir := "./plugins"
	if err == nil {
		defaultPluginDir = appConfig.PluginDir
	}

	// Set default namespace to "default"
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}

	// Check for KUBECONFIG environment variable
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" && len(cfg.KubeconfigPaths) == 0 {
		cfg.KubeconfigPaths = append(cfg.KubeconfigPaths, kubeconfig)
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
				case "session":
					cfg.SessionFile = flagValue
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

	// Handle session loading
	if cfg.SessionFile != "" {
		clusters, err := parseSessionFile(cfg.SessionFile)
		if err != nil {
			fmt.Printf("Warning: Failed to parse session file %s: %v\n", cfg.SessionFile, err)
		} else {
			cfg.SessionClusters = clusters
			// Add kubeconfig paths from session
			for _, cluster := range clusters {
				if cluster.Kubeconfig != "" {
					cfg.KubeconfigPaths = append(cfg.KubeconfigPaths, cluster.Kubeconfig)
				}
			}
			fmt.Printf("Loaded %d clusters from session file: %s\n", len(clusters), cfg.SessionFile)
		}
	}

	// If no kubeconfig specified, use empty string to let k8s client use default behavior
	if len(cfg.KubeconfigPaths) == 0 {
		cfg.KubeconfigPaths = []string{""}
	}

	return cfg
}

func HandlePluginArgs(pluginManager *plugins.GlobalPluginManager, pluginArgs map[string]string) error {
	if pluginManager == nil {
		return nil
	}

	if pluginArgsProcessed {
		fmt.Printf("DEBUG: Plugin args already processed, skipping\n")
		return nil
	}

	api := pluginManager.GetAPI()
	if api == nil {
		return nil
	}

	for argName, argValue := range pluginArgs {
		if api.HasCLIArgument(argName) {
			fmt.Printf("DEBUG: Processing plugin arg: %s = %s\n", argName, argValue)
			if err := api.ExecuteCLIArgument(argName, argValue); err != nil {
				return err
			}
		}
	}

	pluginArgsProcessed = true
	return nil
}
