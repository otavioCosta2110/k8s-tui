package plugins

import (
	"fmt"

	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/yuin/gopher-lua"
)

type LuaUtils struct{}

func NewLuaUtils() *LuaUtils {
	return &LuaUtils{}
}

func (lu *LuaUtils) CallLuaFunction(L *lua.LState, functionName string, args []string) (string, error) {
	if L.GetGlobal(functionName).Type() != lua.LTFunction {
		return "", fmt.Errorf("Lua function %s not found", functionName)
	}

	luaArgs := make([]lua.LValue, len(args))
	for i, arg := range args {
		luaArgs[i] = lua.LString(arg)
	}

	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal(functionName),
		NRet:    2,
		Protect: true,
	}, luaArgs...); err != nil {
		logger.PluginError("lua", fmt.Sprintf("Error calling Lua function %s: %v", functionName, err))
		return "", err
	}

	result := L.Get(-2)
	errorValue := L.Get(-1)
	L.Pop(2)

	if errorValue.Type() == lua.LTString && errorValue.String() != "" {
		return result.String(), fmt.Errorf("%s", errorValue.String())
	}

	return result.String(), nil
}

func (lu *LuaUtils) GetStringField(tbl *lua.LTable, key string) string {
	if val := tbl.RawGetString(key); val.Type() == lua.LTString {
		return val.String()
	}
	return ""
}

func (lu *LuaUtils) ParseTableArray(tbl *lua.LTable) []string {
	var result []string
	tbl.ForEach(func(key, value lua.LValue) {
		if value.Type() == lua.LTString {
			result = append(result, value.String())
		}
	})
	return result
}

func (lu *LuaUtils) SafeCallLuaFunction(L *lua.LState, functionName string, args ...lua.LValue) (lua.LValue, error) {
	fn := L.GetGlobal(functionName)
	if fn.Type() != lua.LTFunction {
		return lua.LNil, fmt.Errorf("function %s not found", functionName)
	}

	if err := L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    1,
		Protect: true,
	}, args...); err != nil {
		return lua.LNil, fmt.Errorf("error calling %s: %v", functionName, err)
	}

	result := L.Get(-1)
	L.Pop(1)
	return result, nil
}

func (lu *LuaUtils) RegisterAPIFunction(L *lua.LState, apiTable *lua.LTable, name string, fn lua.LGFunction) {
	L.SetField(apiTable, name, L.NewFunction(fn))
}

func (lu *LuaUtils) CreateAPITable(L *lua.LState) *lua.LTable {
	return L.NewTable()
}

func (lu *LuaUtils) SetGlobalAPITable(L *lua.LState, apiTable *lua.LTable, name string) {
	L.SetGlobal(name, apiTable)
}
