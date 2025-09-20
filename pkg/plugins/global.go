package plugins

var globalPluginManager *PluginManager

// SetGlobalPluginManager sets the global plugin manager instance
func SetGlobalPluginManager(pm *PluginManager) {
	globalPluginManager = pm
}

// GetGlobalPluginManager returns the global plugin manager instance
func GetGlobalPluginManager() *PluginManager {
	return globalPluginManager
}
