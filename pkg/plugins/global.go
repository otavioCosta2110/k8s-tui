package plugins

var globalPluginManager *GlobalPluginManager

func SetGlobalPluginManager(pm *GlobalPluginManager) {
	globalPluginManager = pm
}

func GetGlobalPluginManager() *GlobalPluginManager {
	return globalPluginManager
}
