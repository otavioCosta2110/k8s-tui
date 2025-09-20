package plugins

var globalPluginManager *PluginManager

func SetGlobalPluginManager(pm *PluginManager) {
	globalPluginManager = pm
}

func GetGlobalPluginManager() *PluginManager {
	return globalPluginManager
}
