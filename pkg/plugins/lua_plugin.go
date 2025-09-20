package plugins

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	"github.com/yuin/gopher-lua"
)

type LuaPlugin struct {
	L          *lua.LState
	pluginName string
}

func (lp *LuaPlugin) Name() string {
	if err := lp.L.CallByParam(lua.P{
		Fn:      lp.L.GetGlobal("Name"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		logger.Error(fmt.Sprintf("🔌 Lua Plugin: Error calling Name(): %v", err))
		return "unknown"
	}
	ret := lp.L.Get(-1)
	lp.L.Pop(1)
	name := ret.String()
	logger.Debug(fmt.Sprintf("🔌 Lua Plugin: Name() returned: %s", name))
	return name
}

func (lp *LuaPlugin) Version() string {
	if lp.L.GetGlobal("Version").Type() == lua.LTFunction {
		if err := lp.L.CallByParam(lua.P{
			Fn:      lp.L.GetGlobal("Version"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			return "1.0.0"
		}
		ret := lp.L.Get(-1)
		lp.L.Pop(1)
		return ret.String()
	}
	return "1.0.0"
}

func (lp *LuaPlugin) Description() string {
	if lp.L.GetGlobal("Description").Type() == lua.LTFunction {
		if err := lp.L.CallByParam(lua.P{
			Fn:      lp.L.GetGlobal("Description"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			return "Lua plugin"
		}
		ret := lp.L.Get(-1)
		lp.L.Pop(1)
		return ret.String()
	}
	return "Lua plugin"
}

func (lp *LuaPlugin) Initialize() error {
	logger.Debug("🔌 Lua Plugin: Calling Initialize()")
	if err := lp.L.CallByParam(lua.P{
		Fn:      lp.L.GetGlobal("Initialize"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		logger.Error(fmt.Sprintf("🔌 Lua Plugin: Error calling Initialize(): %v", err))
		return err
	}
	ret := lp.L.Get(-1)
	lp.L.Pop(1)
	if ret.Type() == lua.LTString {
		errorMsg := ret.String()
		logger.Error(fmt.Sprintf("🔌 Lua Plugin: Initialize() returned error: %s", errorMsg))
		return fmt.Errorf("%s", errorMsg)
	}
	logger.Debug("🔌 Lua Plugin: Initialize() completed successfully")
	return nil
}

func (lp *LuaPlugin) Shutdown() error {
	if lp.L.GetGlobal("Shutdown").Type() == lua.LTFunction {
		if err := lp.L.CallByParam(lua.P{
			Fn:      lp.L.GetGlobal("Shutdown"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			return err
		}
		ret := lp.L.Get(-1)
		lp.L.Pop(1)
		if ret.Type() == lua.LTString {
			return fmt.Errorf("%s", ret.String())
		}
	}
	return nil
}

func (lp *LuaPlugin) hasResourcePlugin() bool {
	return lp.L.GetGlobal("GetResourceTypes").Type() == lua.LTFunction &&
		lp.L.GetGlobal("GetResourceData").Type() == lua.LTFunction
}

func (lp *LuaPlugin) hasUIPlugin() bool {
	return lp.L.GetGlobal("GetUIExtensions").Type() == lua.LTFunction
}

func (lp *LuaPlugin) GetResourceTypes() []CustomResourceType {
	logger.Debug("🔌 Lua Plugin: Calling GetResourceTypes()")
	if err := lp.L.CallByParam(lua.P{
		Fn:      lp.L.GetGlobal("GetResourceTypes"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		logger.Error(fmt.Sprintf("🔌 Lua Plugin: Error calling GetResourceTypes(): %v", err))
		return nil
	}
	ret := lp.L.Get(-1)
	lp.L.Pop(1)

	if ret.Type() != lua.LTTable {
		logger.Error("🔌 Lua Plugin: GetResourceTypes() did not return a table")
		return nil
	}

	var types []CustomResourceType
	tbl := ret.(*lua.LTable)
	tbl.ForEach(func(key, value lua.LValue) {
		if value.Type() == lua.LTTable {
			rt := lp.parseCustomResourceType(value.(*lua.LTable))
			types = append(types, rt)
		}
	})

	logger.Debug(fmt.Sprintf("🔌 Lua Plugin: GetResourceTypes() returned %d resource types", len(types)))
	for i, rt := range types {
		logger.Debug(fmt.Sprintf("🔌 Lua Plugin:   [%d] %s (%s)", i+1, rt.Name, rt.Type))
	}

	return types
}

func (lp *LuaPlugin) parseCustomResourceType(tbl *lua.LTable) CustomResourceType {
	rt := CustomResourceType{}

	rt.Name = lp.getStringField(tbl, "Name")
	rt.Type = lp.getStringField(tbl, "Type")
	rt.Icon = lp.getStringField(tbl, "Icon")
	rt.Namespaced = lp.getBoolField(tbl, "Namespaced")
	refreshSeconds := lp.getNumberField(tbl, "RefreshIntervalSeconds")
	if refreshSeconds <= 0 {
		refreshSeconds = 10 // Default to 10 seconds if not set or invalid
	}
	rt.RefreshInterval = time.Duration(refreshSeconds) * time.Second
	logger.Debug(fmt.Sprintf("🔌 Lua Plugin: RefreshIntervalSeconds = %f, RefreshInterval = %v", refreshSeconds, rt.RefreshInterval))
	rt.Category = lp.getStringField(tbl, "Category")
	rt.Description = lp.getStringField(tbl, "Description")

	logger.Debug(fmt.Sprintf("🔌 Lua Plugin: Parsing resource type - Name: '%s', Type: '%s', Icon: '%s', Namespaced: %t, Refresh: %v, Category: '%s'",
		rt.Name, rt.Type, rt.Icon, rt.Namespaced, rt.RefreshInterval, rt.Category))

	// Parse display component if specified
	if displayComp := lp.getTableField(tbl, "DisplayComponent"); displayComp != nil {
		rt.DisplayComponent = lp.parseDisplayComponent(displayComp)
		logger.Debug(fmt.Sprintf("🔌 Lua Plugin: Display component type: %s", rt.DisplayComponent.Type))
	} else {
		// Legacy support: parse columns for table display
		if cols := lp.getTableField(tbl, "Columns"); cols != nil {
			cols.ForEach(func(_, col lua.LValue) {
				if col.Type() == lua.LTTable {
					column := lp.parseTableColumn(col.(*lua.LTable))
					rt.Columns = append(rt.Columns, column)
				}
			})
			// Set default display component to table
			rt.DisplayComponent = DisplayComponent{
				Type: "table",
				Config: map[string]interface{}{
					"columns": rt.Columns,
				},
			}
		}
	}

	return rt
}

func (lp *LuaPlugin) parseTableColumn(tbl *lua.LTable) table.Column {
	return table.Column{
		Title: lp.getStringField(tbl, "Title"),
		Width: int(lp.getNumberField(tbl, "Width")),
	}
}

func (lp *LuaPlugin) parseDisplayComponent(tbl *lua.LTable) DisplayComponent {
	dc := DisplayComponent{}
	dc.Type = lp.getStringField(tbl, "Type")

	// Parse config table
	if config := lp.getTableField(tbl, "Config"); config != nil {
		dc.Config = make(map[string]interface{})
		config.ForEach(func(key, value lua.LValue) {
			if key.Type() == lua.LTString {
				keyStr := key.String()
				switch value.Type() {
				case lua.LTString:
					dc.Config[keyStr] = value.String()
				case lua.LTNumber:
					dc.Config[keyStr] = float64(value.(lua.LNumber))
				case lua.LTBool:
					dc.Config[keyStr] = lua.LVAsBool(value)
				case lua.LTTable:
					// Parse table as slice of interfaces
					tableSlice := []interface{}{}
					value.(*lua.LTable).ForEach(func(_, item lua.LValue) {
						switch item.Type() {
						case lua.LTNumber:
							tableSlice = append(tableSlice, float64(item.(lua.LNumber)))
						case lua.LTString:
							tableSlice = append(tableSlice, item.String())
						case lua.LTBool:
							tableSlice = append(tableSlice, lua.LVAsBool(item))
							// Add more types if needed
						}
					})
					dc.Config[keyStr] = tableSlice
				}
			}
		})
	}

	// Parse style table
	if style := lp.getTableField(tbl, "Style"); style != nil {
		dc.Style.Width = int(lp.getNumberField(style, "Width"))
		dc.Style.Height = int(lp.getNumberField(style, "Height"))
		dc.Style.Border = lp.getStringField(style, "Border")
		dc.Style.ForegroundColor = lp.getStringField(style, "ForegroundColor")
		dc.Style.BackgroundColor = lp.getStringField(style, "BackgroundColor")
		dc.Style.BorderColor = lp.getStringField(style, "BorderColor")
	}

	return dc
}

func (lp *LuaPlugin) GetResourceData(client k8s.Client, resourceType string, namespace string) ([]types.ResourceData, error) {
	logger.PluginDebug(lp.pluginName, fmt.Sprintf("Calling GetResourceData(%s, %s)", resourceType, namespace))

	if err := lp.L.CallByParam(lua.P{
		Fn:      lp.L.GetGlobal("GetResourceData"),
		NRet:    1,
		Protect: true,
	}, lua.LString(resourceType), lua.LString(namespace)); err != nil {
		logger.PluginError(lp.pluginName, fmt.Sprintf("Error calling GetResourceData(): %v", err))
		return nil, err
	}
	ret := lp.L.Get(-1)
	lp.L.Pop(1)

	if ret.Type() != lua.LTTable {
		logger.PluginError(lp.pluginName, "GetResourceData() did not return a table")
		return nil, fmt.Errorf("GetResourceData must return a table")
	}

	var data []types.ResourceData
	tbl := ret.(*lua.LTable)
	tbl.ForEach(func(_, item lua.LValue) {
		if item.Type() == lua.LTTable {
			rd := lp.parseResourceData(item.(*lua.LTable))
			data = append(data, rd)
		}
	})

	logger.PluginDebug(lp.pluginName, fmt.Sprintf("GetResourceData() returned %d data items", len(data)))
	return data, nil
}

func (lp *LuaPlugin) parseResourceData(tbl *lua.LTable) types.ResourceData {
	// Extract standard fields
	name := lp.getStringField(tbl, "Name")
	namespace := lp.getStringField(tbl, "Namespace")
	status := lp.getStringField(tbl, "Status")
	age := lp.getStringField(tbl, "Age")

	// Extract all fields dynamically for custom plugins
	fields := make(map[string]string)
	tbl.ForEach(func(key lua.LValue, value lua.LValue) {
		if key.Type() == lua.LTString {
			fieldName := key.String()
			fieldValue := ""
			if value.Type() == lua.LTString {
				fieldValue = value.String()
			}
			fields[fieldName] = fieldValue
			logger.PluginDebug(lp.pluginName, fmt.Sprintf("Extracted field '%s' = '%s'", fieldName, fieldValue))
		}
	})

	return &LuaResourceData{
		name:      name,
		namespace: namespace,
		status:    status,
		age:       age,
		fields:    fields,
	}
}

func (lp *LuaPlugin) DeleteResource(client k8s.Client, resourceType string, namespace string, name string) error {
	logger.Debug(fmt.Sprintf("🔌 Lua Plugin: Calling DeleteResource(%s, %s, %s)", resourceType, namespace, name))

	if lp.L.GetGlobal("DeleteResource").Type() != lua.LTFunction {
		logger.Error("🔌 Lua Plugin: DeleteResource function not defined")
		return fmt.Errorf("DeleteResource function not defined")
	}
	if err := lp.L.CallByParam(lua.P{
		Fn:      lp.L.GetGlobal("DeleteResource"),
		NRet:    1,
		Protect: true,
	}, lua.LString(resourceType), lua.LString(namespace), lua.LString(name)); err != nil {
		logger.Error(fmt.Sprintf("🔌 Lua Plugin: Error calling DeleteResource(): %v", err))
		return err
	}
	ret := lp.L.Get(-1)
	lp.L.Pop(1)
	if ret.Type() == lua.LTString {
		errorMsg := ret.String()
		logger.Error(fmt.Sprintf("🔌 Lua Plugin: DeleteResource() returned error: %s", errorMsg))
		return fmt.Errorf("%s", errorMsg)
	}

	logger.Debug("🔌 Lua Plugin: DeleteResource() completed successfully")
	return nil
}

func (lp *LuaPlugin) GetResourceInfo(client k8s.Client, resourceType string, namespace string, name string) (*k8s.ResourceInfo, error) {
	logger.Debug(fmt.Sprintf("🔌 Lua Plugin: Calling GetResourceInfo(%s, %s, %s)", resourceType, namespace, name))

	if lp.L.GetGlobal("GetResourceInfo").Type() != lua.LTFunction {
		logger.Error("🔌 Lua Plugin: GetResourceInfo function not defined")
		return nil, fmt.Errorf("GetResourceInfo function not defined")
	}
	if err := lp.L.CallByParam(lua.P{
		Fn:      lp.L.GetGlobal("GetResourceInfo"),
		NRet:    1,
		Protect: true,
	}, lua.LString(resourceType), lua.LString(namespace), lua.LString(name)); err != nil {
		logger.Error(fmt.Sprintf("🔌 Lua Plugin: Error calling GetResourceInfo(): %v", err))
		return nil, err
	}
	ret := lp.L.Get(-1)
	lp.L.Pop(1)

	if ret.Type() != lua.LTTable {
		logger.Error("🔌 Lua Plugin: GetResourceInfo() did not return a table")
		return nil, fmt.Errorf("GetResourceInfo must return a table")
	}

	tbl := ret.(*lua.LTable)
	resourceInfo := &k8s.ResourceInfo{
		Name:      lp.getStringField(tbl, "Name"),
		Namespace: lp.getStringField(tbl, "Namespace"),
		Kind:      k8s.ResourceType(lp.getStringField(tbl, "Kind")),
		Age:       lp.getStringField(tbl, "Age"),
	}

	logger.Debug(fmt.Sprintf("🔌 Lua Plugin: GetResourceInfo() returned: %s/%s (%s)", resourceInfo.Namespace, resourceInfo.Name, resourceInfo.Kind))
	return resourceInfo, nil
}

func (lp *LuaPlugin) GetUIExtensions() []UIExtension {
	if lp.L.GetGlobal("GetUIExtensions").Type() != lua.LTFunction {
		return nil
	}
	if err := lp.L.CallByParam(lua.P{
		Fn:      lp.L.GetGlobal("GetUIExtensions"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		return nil
	}
	ret := lp.L.Get(-1)
	lp.L.Pop(1)

	if ret.Type() != lua.LTTable {
		return nil
	}

	var extensions []UIExtension
	tbl := ret.(*lua.LTable)
	tbl.ForEach(func(_, ext lua.LValue) {
		if ext.Type() == lua.LTTable {
			extension := lp.parseUIExtension(ext.(*lua.LTable))
			extensions = append(extensions, extension)
		}
	})
	return extensions
}

func (lp *LuaPlugin) parseUIExtension(tbl *lua.LTable) UIExtension {
	ext := UIExtension{
		Name:       lp.getStringField(tbl, "Name"),
		Type:       lp.getStringField(tbl, "Type"),
		KeyBinding: lp.getStringField(tbl, "KeyBinding"),
	}

	// Parse injection points
	if injectionPoints := lp.getTableField(tbl, "InjectionPoints"); injectionPoints != nil {
		injectionPoints.ForEach(func(_, point lua.LValue) {
			if point.Type() == lua.LTTable {
				ip := lp.parseUIInjectionPoint(point.(*lua.LTable))
				ext.InjectionPoints = append(ext.InjectionPoints, ip)
			}
		})
	}

	// Parse interactions
	if interactions := lp.getTableField(tbl, "Interactions"); interactions != nil {
		interactions.ForEach(func(_, interaction lua.LValue) {
			if interaction.Type() == lua.LTTable {
				interact := lp.parseInteraction(interaction.(*lua.LTable))
				ext.Interactions = append(ext.Interactions, interact)
			}
		})
	}

	// Parse dependencies
	if dependencies := lp.getTableField(tbl, "Dependencies"); dependencies != nil {
		dependencies.ForEach(func(_, dep lua.LValue) {
			if dep.Type() == lua.LTString {
				ext.Dependencies = append(ext.Dependencies, dep.String())
			}
		})
	}

	return ext
}

func (lp *LuaPlugin) parseUIInjectionPoint(tbl *lua.LTable) UIInjectionPoint {
	ip := UIInjectionPoint{}
	ip.Location = lp.getStringField(tbl, "Location")
	ip.Position = lp.getStringField(tbl, "Position")
	ip.Priority = int(lp.getNumberField(tbl, "Priority"))
	ip.DataSource = lp.getStringField(tbl, "DataSource")
	ip.UpdateInterval = int(lp.getNumberField(tbl, "UpdateInterval"))

	if component := lp.getTableField(tbl, "Component"); component != nil {
		ip.Component = lp.parseDisplayComponent(component)
	}

	return ip
}

func (lp *LuaPlugin) parseInteraction(tbl *lua.LTable) Interaction {
	return Interaction{
		Type:       lp.getStringField(tbl, "Type"),
		Label:      lp.getStringField(tbl, "Label"),
		KeyBinding: lp.getStringField(tbl, "KeyBinding"),
		Context:    lp.getStringField(tbl, "Context"),
		Enabled:    lp.getBoolField(tbl, "Enabled"),
		Tooltip:    lp.getStringField(tbl, "Tooltip"),
	}
}

func (lp *LuaPlugin) getStringField(tbl *lua.LTable, key string) string {
	val := tbl.RawGetString(key)
	if val.Type() == lua.LTString {
		result := val.String()
		logger.Debug(fmt.Sprintf("🔌 Lua Plugin: getStringField('%s') = '%s'", key, result))
		return result
	}
	logger.Debug(fmt.Sprintf("🔌 Lua Plugin: getStringField('%s') = empty (type: %s)", key, val.Type().String()))
	return ""
}

func (lp *LuaPlugin) getBoolField(tbl *lua.LTable, key string) bool {
	if val := tbl.RawGetString(key); val.Type() == lua.LTBool {
		return lua.LVAsBool(val)
	}
	return false
}

func (lp *LuaPlugin) getNumberField(tbl *lua.LTable, key string) float64 {
	if val := tbl.RawGetString(key); val.Type() == lua.LTNumber {
		result := float64(val.(lua.LNumber))
		logger.Debug(fmt.Sprintf("🔌 Lua Plugin: getNumberField('%s') = %f", key, result))
		return result
	}
	logger.Debug(fmt.Sprintf("🔌 Lua Plugin: getNumberField('%s') not found or not a number, returning 0", key))
	return 0
}

func (lp *LuaPlugin) getTableField(tbl *lua.LTable, key string) *lua.LTable {
	if val := tbl.RawGetString(key); val.Type() == lua.LTTable {
		return val.(*lua.LTable)
	}
	return nil
}

type LuaResourceData struct {
	name      string
	namespace string
	status    string
	age       string
	fields    map[string]string
}

func (lrd *LuaResourceData) GetName() string {
	return lrd.name
}

func (lrd *LuaResourceData) GetNamespace() string {
	return lrd.namespace
}

func (lrd *LuaResourceData) GetColumns() table.Row {
	// Return columns in the expected order: Name, Namespace, Status, Age
	// This matches the table model definition in NewCustomResourceTableModel
	return table.Row{lrd.name, lrd.namespace, lrd.status, lrd.age}
}

func (lrd *LuaResourceData) GetFields() map[string]string {
	return lrd.fields
}
