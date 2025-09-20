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
	L *lua.LState
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
			return fmt.Errorf(ret.String())
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
	rt.RefreshInterval = time.Duration(lp.getNumberField(tbl, "RefreshIntervalSeconds")) * time.Second

	logger.Debug(fmt.Sprintf("🔌 Lua Plugin: Parsing resource type - Name: '%s', Type: '%s', Icon: '%s', Namespaced: %t, Refresh: %v",
		rt.Name, rt.Type, rt.Icon, rt.Namespaced, rt.RefreshInterval))

	if cols := lp.getTableField(tbl, "Columns"); cols != nil {
		cols.ForEach(func(_, col lua.LValue) {
			if col.Type() == lua.LTTable {
				column := lp.parseTableColumn(col.(*lua.LTable))
				rt.Columns = append(rt.Columns, column)
			}
		})
	}

	return rt
}

func (lp *LuaPlugin) parseTableColumn(tbl *lua.LTable) table.Column {
	return table.Column{
		Title: lp.getStringField(tbl, "Title"),
		Width: int(lp.getNumberField(tbl, "Width")),
	}
}

func (lp *LuaPlugin) GetResourceData(client k8s.Client, resourceType string, namespace string) ([]types.ResourceData, error) {
	logger.Debug(fmt.Sprintf("🔌 Lua Plugin: Calling GetResourceData(%s, %s)", resourceType, namespace))

	if err := lp.L.CallByParam(lua.P{
		Fn:      lp.L.GetGlobal("GetResourceData"),
		NRet:    1,
		Protect: true,
	}, lua.LString(resourceType), lua.LString(namespace)); err != nil {
		logger.Error(fmt.Sprintf("🔌 Lua Plugin: Error calling GetResourceData(): %v", err))
		return nil, err
	}
	ret := lp.L.Get(-1)
	lp.L.Pop(1)

	if ret.Type() != lua.LTTable {
		logger.Error("🔌 Lua Plugin: GetResourceData() did not return a table")
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

	logger.Debug(fmt.Sprintf("🔌 Lua Plugin: GetResourceData() returned %d data items", len(data)))
	return data, nil
}

func (lp *LuaPlugin) parseResourceData(tbl *lua.LTable) types.ResourceData {
	return &LuaResourceData{
		name:      lp.getStringField(tbl, "Name"),
		namespace: lp.getStringField(tbl, "Namespace"),
		status:    lp.getStringField(tbl, "Status"),
		age:       lp.getStringField(tbl, "Age"),
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
		return fmt.Errorf(errorMsg)
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
	return UIExtension{
		Name:       lp.getStringField(tbl, "Name"),
		Type:       lp.getStringField(tbl, "Type"),
		KeyBinding: lp.getStringField(tbl, "KeyBinding"),
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
		return float64(val.(lua.LNumber))
	}
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
}

func (lrd *LuaResourceData) GetName() string {
	return lrd.name
}

func (lrd *LuaResourceData) GetNamespace() string {
	return lrd.namespace
}

func (lrd *LuaResourceData) GetColumns() table.Row {
	return table.Row{lrd.name, lrd.namespace, lrd.status, lrd.age}
}
