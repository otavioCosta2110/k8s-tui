package ui

func (m *AppModel) loadPluginUIExtensions() {
	if m.pluginManager == nil {
		return
	}

	registry := m.pluginManager.GetRegistry()
	if registry == nil {
		return
	}

	for _, plugin := range registry.GetUIPlugins() {
		extensions := plugin.GetUIExtensions()
		for _, ext := range extensions {
			for _, injection := range ext.InjectionPoints {
				m.uiInjector.AddInjection(injection.Location, injection)
			}
		}
	}

	api := m.pluginManager.GetAPI()
	if api != nil {
		headerComponents := api.GetHeaderComponents()
		for _, component := range headerComponents {
			if content, ok := component.Component.Config["content"].(string); ok {
				m.header.AddPluginComponent(content)
			}
		}

		footerComponents := api.GetFooterComponents()
		for _, component := range footerComponents {
			if _, ok := component.Component.Config["content"].(string); ok {
				m.uiInjector.AddInjection("footer", component)
			}
		}
	}
}
