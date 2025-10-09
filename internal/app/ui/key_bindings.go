package ui

func (m *AppModel) getKeyBinding(action string) string {
	if binding, exists := m.config.KeyBindings[action]; exists {
		return binding
	}
	defaults := map[string]string{
		"quit":      "q",
		"help":      "?",
		"refresh":   "r",
		"back":      "[",
		"forward":   "]",
		"new_tab":   "ctrl+t",
		"close_tab": "ctrl+w",
		"quick_nav": "g",
	}
	return defaults[action]
}

func (m *AppModel) getResourceTypeFromKey(key string) string {
	resourceMap := map[string]string{
		"p": "Pods",
		"d": "Deployments",
		"s": "Services",
		"i": "Ingresses",
		"c": "ConfigMaps",
		"e": "Secrets",
		"a": "ServiceAccounts",
		"r": "ReplicaSets",
		"n": "Nodes",
		"j": "Jobs",
		"k": "CronJobs",
		"m": "DaemonSets",
		"t": "StatefulSets",
		"l": "ResourceList",
	}

	if resourceType, exists := resourceMap[key]; exists {
		return resourceType
	}
	return ""
}
