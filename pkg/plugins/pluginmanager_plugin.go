package plugins

import (
	"context"
	"fmt"
	"github.com/otavioCosta2110/k8s-tui/pkg/format"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/yuin/gopher-lua"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type PluginmanagerStyleLuaPlugin struct {
	L          *lua.LState
	pluginName string
	config     map[string]interface{}
	api        PluginAPI
}

func NewPluginmanagerStyleLuaPlugin(L *lua.LState, pluginName string, api PluginAPI) *PluginmanagerStyleLuaPlugin {
	return &PluginmanagerStyleLuaPlugin{
		L:          L,
		pluginName: pluginName,
		config:     make(map[string]interface{}),
		api:        api,
	}
}

func (p *PluginmanagerStyleLuaPlugin) Name() string {
	if err := p.L.CallByParam(lua.P{
		Fn:      p.L.GetGlobal("Name"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		logger.PluginError(p.pluginName, fmt.Sprintf("Error calling Name(): %v", err))
		return p.pluginName
	}
	ret := p.L.Get(-1)
	p.L.Pop(1)
	return ret.String()
}

func (p *PluginmanagerStyleLuaPlugin) Version() string {
	if p.L.GetGlobal("Version").Type() == lua.LTFunction {
		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Version"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			return "1.0.0"
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		return ret.String()
	}
	return "1.0.0"
}

func (p *PluginmanagerStyleLuaPlugin) Description() string {
	if p.L.GetGlobal("Description").Type() == lua.LTFunction {
		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Description"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			return "Pluginmanager-style Lua plugin"
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		return ret.String()
	}
	return "Neovim-style Lua plugin"
}

func (p *PluginmanagerStyleLuaPlugin) Initialize() error {
	logger.PluginDebug(p.pluginName, "Initializing pluginmanager-style plugin")

	// Set up the plugin API in Lua
	p.setupLuaAPI()

	if err := p.L.CallByParam(lua.P{
		Fn:      p.L.GetGlobal("Initialize"),
		NRet:    1,
		Protect: true,
	}); err != nil {
		logger.PluginError(p.pluginName, fmt.Sprintf("Error calling Initialize(): %v", err))
		return err
	}
	ret := p.L.Get(-1)
	p.L.Pop(1)
	if ret.Type() == lua.LTString {
		errorMsg := ret.String()
		logger.PluginError(p.pluginName, fmt.Sprintf("Initialize() returned error: %s", errorMsg))
		return fmt.Errorf("%s", errorMsg)
	}
	return nil
}

func (p *PluginmanagerStyleLuaPlugin) Shutdown() error {
	if p.L.GetGlobal("Shutdown").Type() == lua.LTFunction {
		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Shutdown"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			return err
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		if ret.Type() == lua.LTString {
			return fmt.Errorf("%s", ret.String())
		}
	}
	return nil
}

func (p *PluginmanagerStyleLuaPlugin) Setup(opts map[string]interface{}) error {
	logger.PluginDebug(p.pluginName, "Setting up plugin with options")

	// Merge options with existing config
	for k, v := range opts {
		p.config[k] = v
	}

	// Call Lua setup function if it exists
	setupType := p.L.GetGlobal("Setup").Type()
	logger.PluginDebug(p.pluginName, fmt.Sprintf("Setup function type: %s", setupType))

	if setupType == lua.LTFunction {
		// Convert Go map to Lua table
		optsTable := p.L.NewTable()
		for k, v := range opts {
			optsTable.RawSetString(k, lua.LString(fmt.Sprintf("%v", v)))
		}

		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Setup"),
			NRet:    1,
			Protect: true,
		}, optsTable); err != nil {
			logger.PluginError(p.pluginName, fmt.Sprintf("Error calling Setup(): %v", err))
			return err
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		if ret.Type() == lua.LTString {
			return fmt.Errorf("%s", ret.String())
		}
	}

	return nil
}

func (p *PluginmanagerStyleLuaPlugin) Config() map[string]interface{} {
	configType := p.L.GetGlobal("Config").Type()
	logger.PluginDebug(p.pluginName, fmt.Sprintf("Config function type: %s", configType))

	if configType == lua.LTFunction {
		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Config"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			logger.PluginError(p.pluginName, fmt.Sprintf("Error calling Config(): %v", err))
			return p.config
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		if ret.Type() == lua.LTTable {
			// Convert Lua table to Go map
			config := make(map[string]interface{})
			ret.(*lua.LTable).ForEach(func(key, value lua.LValue) {
				if key.Type() == lua.LTString {
					config[key.String()] = value.String()
				}
			})
			return config
		}
	}
	return p.config
}

func (p *PluginmanagerStyleLuaPlugin) Commands() []PluginCommand {
	var commands []PluginCommand

	if p.L.GetGlobal("Commands").Type() == lua.LTFunction {
		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Commands"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			logger.PluginError(p.pluginName, fmt.Sprintf("Error calling Commands(): %v", err))
			return commands
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		if ret.Type() == lua.LTTable {
			ret.(*lua.LTable).ForEach(func(key, value lua.LValue) {
				if value.Type() == lua.LTTable {
					cmd := p.parsePluginCommand(value.(*lua.LTable))
					commands = append(commands, cmd)
				}
			})
		}
	}

	return commands
}

func (p *PluginmanagerStyleLuaPlugin) Hooks() []PluginHook {
	var hooks []PluginHook

	if p.L.GetGlobal("Hooks").Type() == lua.LTFunction {
		if err := p.L.CallByParam(lua.P{
			Fn:      p.L.GetGlobal("Hooks"),
			NRet:    1,
			Protect: true,
		}); err != nil {
			logger.PluginError(p.pluginName, fmt.Sprintf("Error calling Hooks(): %v", err))
			return hooks
		}
		ret := p.L.Get(-1)
		p.L.Pop(1)
		if ret.Type() == lua.LTTable {
			ret.(*lua.LTable).ForEach(func(key, value lua.LValue) {
				if value.Type() == lua.LTTable {
					hook := p.parsePluginHook(value.(*lua.LTable))
					hooks = append(hooks, hook)
				}
			})
		}
	}

	return hooks
}

func (p *PluginmanagerStyleLuaPlugin) parsePluginCommand(tbl *lua.LTable) PluginCommand {
	return PluginCommand{
		Name:        p.getStringField(tbl, "name"),
		Description: p.getStringField(tbl, "description"),
		Handler: func(args []string) (string, error) {
			// This would need to be implemented to call Lua functions
			return "Command executed", nil
		},
	}
}

func (p *PluginmanagerStyleLuaPlugin) parsePluginHook(tbl *lua.LTable) PluginHook {
	event := p.getStringField(tbl, "event")
	handlerName := p.getStringField(tbl, "handler")

	return PluginHook{
		Event: event,
		Handler: func(data interface{}) error {
			// Call the Lua handler function
			if p.L.GetGlobal(handlerName).Type() == lua.LTFunction {
				if err := p.L.CallByParam(lua.P{
					Fn:      p.L.GetGlobal(handlerName),
					NRet:    1,
					Protect: true,
				}, lua.LString(fmt.Sprintf("%v", data))); err != nil {
					logger.PluginError(p.pluginName, fmt.Sprintf("Error calling hook handler %s: %v", handlerName, err))
					return err
				}
				ret := p.L.Get(-1)
				p.L.Pop(1)
				if ret.Type() == lua.LTString {
					errorMsg := ret.String()
					logger.PluginError(p.pluginName, fmt.Sprintf("Hook handler %s returned error: %s", handlerName, errorMsg))
					return fmt.Errorf("%s", errorMsg)
				}
				logger.PluginDebug(p.pluginName, fmt.Sprintf("Hook handler %s called successfully", handlerName))
			} else {
				logger.PluginWarn(p.pluginName, fmt.Sprintf("Hook handler function %s not found", handlerName))
			}
			return nil
		},
	}
}

func (p *PluginmanagerStyleLuaPlugin) getStringField(tbl *lua.LTable, key string) string {
	if val := tbl.RawGetString(key); val.Type() == lua.LTString {
		return val.String()
	}
	return ""
}

func (p *PluginmanagerStyleLuaPlugin) setupLuaAPI() {
	// Create the k8s_tui API table
	apiTable := p.L.NewTable()

	// Add API functions
	p.L.SetField(apiTable, "get_namespace", p.L.NewFunction(p.luaGetNamespace))
	p.L.SetField(apiTable, "set_status", p.L.NewFunction(p.luaSetStatus))
	p.L.SetField(apiTable, "add_header", p.L.NewFunction(p.luaAddHeader))
	p.L.SetField(apiTable, "register_command", p.L.NewFunction(p.luaRegisterCommand))

	// Kubernetes resource API functions
	p.L.SetField(apiTable, "get_pods", p.L.NewFunction(p.luaGetPods))
	p.L.SetField(apiTable, "get_services", p.L.NewFunction(p.luaGetServices))
	p.L.SetField(apiTable, "get_deployments", p.L.NewFunction(p.luaGetDeployments))
	p.L.SetField(apiTable, "get_configmaps", p.L.NewFunction(p.luaGetConfigMaps))
	p.L.SetField(apiTable, "get_secrets", p.L.NewFunction(p.luaGetSecrets))
	p.L.SetField(apiTable, "get_ingresses", p.L.NewFunction(p.luaGetIngresses))
	p.L.SetField(apiTable, "get_jobs", p.L.NewFunction(p.luaGetJobs))
	p.L.SetField(apiTable, "get_cronjobs", p.L.NewFunction(p.luaGetCronJobs))
	p.L.SetField(apiTable, "get_daemonsets", p.L.NewFunction(p.luaGetDaemonSets))
	p.L.SetField(apiTable, "get_statefulsets", p.L.NewFunction(p.luaGetStatefulSets))
	p.L.SetField(apiTable, "get_replicasets", p.L.NewFunction(p.luaGetReplicaSets))
	p.L.SetField(apiTable, "get_nodes", p.L.NewFunction(p.luaGetNodes))
	p.L.SetField(apiTable, "get_namespaces", p.L.NewFunction(p.luaGetNamespaces))
	p.L.SetField(apiTable, "get_serviceaccounts", p.L.NewFunction(p.luaGetServiceAccounts))

	// Delete functions
	p.L.SetField(apiTable, "delete_pod", p.L.NewFunction(p.luaDeletePod))
	p.L.SetField(apiTable, "delete_service", p.L.NewFunction(p.luaDeleteService))
	p.L.SetField(apiTable, "delete_deployment", p.L.NewFunction(p.luaDeleteDeployment))
	p.L.SetField(apiTable, "delete_configmap", p.L.NewFunction(p.luaDeleteConfigMap))
	p.L.SetField(apiTable, "delete_secret", p.L.NewFunction(p.luaDeleteSecret))
	p.L.SetField(apiTable, "delete_ingress", p.L.NewFunction(p.luaDeleteIngress))
	p.L.SetField(apiTable, "delete_job", p.L.NewFunction(p.luaDeleteJob))
	p.L.SetField(apiTable, "delete_cronjob", p.L.NewFunction(p.luaDeleteCronJob))
	p.L.SetField(apiTable, "delete_daemonset", p.L.NewFunction(p.luaDeleteDaemonSet))
	p.L.SetField(apiTable, "delete_statefulset", p.L.NewFunction(p.luaDeleteStatefulSet))
	p.L.SetField(apiTable, "delete_replicaset", p.L.NewFunction(p.luaDeleteReplicaSet))
	p.L.SetField(apiTable, "delete_serviceaccount", p.L.NewFunction(p.luaDeleteServiceAccount))

	p.L.SetField(apiTable, "get_endpoints", p.L.NewFunction(p.luaGetEndpoints))

	// Set the API in the global environment
	p.L.SetGlobal("k8s_tui", apiTable)
}

func (p *PluginmanagerStyleLuaPlugin) luaGetNamespace(L *lua.LState) int {
	namespace := p.api.GetCurrentNamespace()
	L.Push(lua.LString(namespace))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaSetStatus(L *lua.LState) int {
	message := L.CheckString(1)
	p.api.SetStatusMessage(message)
	return 0
}

func (p *PluginmanagerStyleLuaPlugin) luaAddHeader(L *lua.LState) int {
	content := L.CheckString(1)
	component := UIInjectionPoint{
		Location: "header",
		Position: "right",
		Priority: 10,
		Component: DisplayComponent{
			Type: "text",
			Config: map[string]any{
				"content": content,
				"style":   "info",
			},
		},
		DataSource:     "static",
		UpdateInterval: 0,
	}
	p.api.AddHeaderComponent(component)
	return 0
}

func (p *PluginmanagerStyleLuaPlugin) luaRegisterCommand(L *lua.LState) int {
	name := L.CheckString(1)
	description := L.CheckString(2)
	// Store the command for later execution
	command := PluginCommand{
		Name:        name,
		Description: description,
		Handler: func(args []string) (string, error) {
			// This would call the Lua function
			return "Command executed from Lua", nil
		},
	}
	p.api.RegisterCommand(command.Name, command.Description, command.Handler)
	return 0
}

// Kubernetes resource API functions
func (p *PluginmanagerStyleLuaPlugin) luaGetPods(L *lua.LState) int {
	namespace := L.CheckString(1)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	pods, err := k8s.FetchPods(client, namespace, "")
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch pods: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, pod := range pods {
		podTable := L.NewTable()
		L.SetField(podTable, "Name", lua.LString(pod.Name))
		L.SetField(podTable, "Namespace", lua.LString(pod.Namespace))
		L.SetField(podTable, "Ready", lua.LString(pod.Ready))
		L.SetField(podTable, "Status", lua.LString(pod.Status))
		L.SetField(podTable, "Restarts", lua.LNumber(pod.Restarts))
		L.SetField(podTable, "Age", lua.LString(pod.Age))
		L.RawSetInt(resultTable, i+1, podTable)
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetServices(L *lua.LState) int {
	namespace := L.CheckString(1)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	services, err := k8s.GetServicesTableData(client, namespace)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch services: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, svc := range services {
		svcTable := L.NewTable()
		L.SetField(svcTable, "Name", lua.LString(svc.Name))
		L.SetField(svcTable, "Namespace", lua.LString(svc.Namespace))
		L.SetField(svcTable, "Type", lua.LString(svc.Type))
		L.SetField(svcTable, "ClusterIP", lua.LString(svc.ClusterIP))
		L.SetField(svcTable, "ExternalIP", lua.LString(svc.ExternalIP))
		L.SetField(svcTable, "Ports", lua.LString(svc.Ports))
		L.SetField(svcTable, "Age", lua.LString(svc.Age))
		L.RawSetInt(resultTable, i+1, svcTable)
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetDeployments(L *lua.LState) int {
	namespace := L.CheckString(1)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	deployments, err := k8s.GetDeploymentsTableData(client, namespace)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch deployments: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, dep := range deployments {
		depTable := L.NewTable()
		L.SetField(depTable, "Name", lua.LString(dep.Name))
		L.SetField(depTable, "Namespace", lua.LString(dep.Namespace))
		L.SetField(depTable, "Ready", lua.LString(dep.Ready))
		L.SetField(depTable, "UpToDate", lua.LString(dep.UpToDate))
		L.SetField(depTable, "Available", lua.LString(dep.Available))
		L.SetField(depTable, "Age", lua.LString(dep.Age))
		L.RawSetInt(resultTable, i+1, depTable)
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetConfigMaps(L *lua.LState) int {
	namespace := L.CheckString(1)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	configmaps, err := k8s.FetchConfigmaps(client, namespace, "")
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch configmaps: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, cm := range configmaps {
		cmTable := L.NewTable()
		L.SetField(cmTable, "Name", lua.LString(cm.Name))
		L.SetField(cmTable, "Namespace", lua.LString(cm.Namespace))
		L.SetField(cmTable, "Data", lua.LNumber(len(cm.Data)))
		L.SetField(cmTable, "Age", lua.LString(cm.Age))
		L.RawSetInt(resultTable, i+1, cmTable)
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetSecrets(L *lua.LState) int {
	namespace := L.CheckString(1)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	secrets, err := k8s.GetSecretsTableData(client, namespace)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch secrets: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, secret := range secrets {
		secretTable := L.NewTable()
		L.SetField(secretTable, "Name", lua.LString(secret.Name))
		L.SetField(secretTable, "Namespace", lua.LString(secret.Namespace))
		L.SetField(secretTable, "Type", lua.LString(secret.Type))
		L.SetField(secretTable, "Data", lua.LNumber(len(secret.Data)))
		L.SetField(secretTable, "Age", lua.LString(secret.Age))
		L.RawSetInt(resultTable, i+1, secretTable)
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetIngresses(L *lua.LState) int {
	namespace := L.CheckString(1)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	ingresses, err := k8s.GetIngressesTableData(client, namespace)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch ingresses: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, ing := range ingresses {
		ingTable := L.NewTable()
		L.SetField(ingTable, "Name", lua.LString(ing.Name))
		L.SetField(ingTable, "Namespace", lua.LString(ing.Namespace))
		L.SetField(ingTable, "Class", lua.LString(ing.Class))
		L.SetField(ingTable, "Hosts", lua.LString(ing.Hosts))
		L.SetField(ingTable, "Address", lua.LString(ing.Address))
		L.SetField(ingTable, "Ports", lua.LString(ing.Ports))
		L.SetField(ingTable, "Age", lua.LString(ing.Age))
		L.RawSetInt(resultTable, i+1, ingTable)
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetJobs(L *lua.LState) int {
	namespace := L.CheckString(1)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	jobs, err := k8s.GetJobsTableData(client, namespace)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch jobs: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, job := range jobs {
		jobTable := L.NewTable()
		L.SetField(jobTable, "Name", lua.LString(job.Name))
		L.SetField(jobTable, "Namespace", lua.LString(job.Namespace))
		L.SetField(jobTable, "Completions", lua.LString(job.Completions))
		L.SetField(jobTable, "Duration", lua.LString(job.Duration))
		L.SetField(jobTable, "Age", lua.LString(job.Age))
		L.RawSetInt(resultTable, i+1, jobTable)
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetCronJobs(L *lua.LState) int {
	namespace := L.CheckString(1)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	cronjobs, err := k8s.GetCronJobsTableData(client, namespace)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch cronjobs: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, cj := range cronjobs {
		cjTable := L.NewTable()
		L.SetField(cjTable, "Name", lua.LString(cj.Name))
		L.SetField(cjTable, "Namespace", lua.LString(cj.Namespace))
		L.SetField(cjTable, "Schedule", lua.LString(cj.Schedule))
		L.SetField(cjTable, "Suspend", lua.LString(cj.Suspend))
		L.SetField(cjTable, "Active", lua.LString(cj.Active))
		L.SetField(cjTable, "LastSchedule", lua.LString(cj.LastSchedule))
		L.SetField(cjTable, "Age", lua.LString(cj.Age))
		L.RawSetInt(resultTable, i+1, cjTable)
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetDaemonSets(L *lua.LState) int {
	namespace := L.CheckString(1)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	daemonsets, err := k8s.GetDaemonSetsTableData(client, namespace)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch daemonsets: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, ds := range daemonsets {
		dsTable := L.NewTable()
		L.SetField(dsTable, "Name", lua.LString(ds.Name))
		L.SetField(dsTable, "Namespace", lua.LString(ds.Namespace))
		L.SetField(dsTable, "Desired", lua.LString(ds.Desired))
		L.SetField(dsTable, "Current", lua.LString(ds.Current))
		L.SetField(dsTable, "Ready", lua.LString(ds.Ready))
		L.SetField(dsTable, "UpToDate", lua.LString(ds.UpToDate))
		L.SetField(dsTable, "Available", lua.LString(ds.Available))
		L.SetField(dsTable, "Age", lua.LString(ds.Age))
		L.RawSetInt(resultTable, i+1, dsTable)
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetStatefulSets(L *lua.LState) int {
	namespace := L.CheckString(1)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	statefulsets, err := k8s.GetStatefulSetsTableData(client, namespace)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch statefulsets: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, sts := range statefulsets {
		stsTable := L.NewTable()
		L.SetField(stsTable, "Name", lua.LString(sts.Name))
		L.SetField(stsTable, "Namespace", lua.LString(sts.Namespace))
		L.SetField(stsTable, "Ready", lua.LString(sts.Ready))
		L.SetField(stsTable, "Age", lua.LString(sts.Age))
		L.RawSetInt(resultTable, i+1, stsTable)
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetReplicaSets(L *lua.LState) int {
	namespace := L.CheckString(1)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	replicasets, err := k8s.GetReplicaSetsTableData(client, namespace)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch replicasets: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, rs := range replicasets {
		rsTable := L.NewTable()
		L.SetField(rsTable, "Name", lua.LString(rs.Name))
		L.SetField(rsTable, "Namespace", lua.LString(rs.Namespace))
		L.SetField(rsTable, "Desired", lua.LString(rs.Desired))
		L.SetField(rsTable, "Current", lua.LString(rs.Current))
		L.SetField(rsTable, "Ready", lua.LString(rs.Ready))
		L.SetField(rsTable, "Age", lua.LString(rs.Age))
		L.RawSetInt(resultTable, i+1, rsTable)
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetNodes(L *lua.LState) int {
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	nodes, err := k8s.GetNodesTableData(client)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch nodes: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, node := range nodes {
		nodeTable := L.NewTable()
		L.SetField(nodeTable, "Name", lua.LString(node.Name))
		L.SetField(nodeTable, "Status", lua.LString(node.Status))
		L.SetField(nodeTable, "Roles", lua.LString(node.Roles))
		L.SetField(nodeTable, "Age", lua.LString(node.Age))
		L.SetField(nodeTable, "Version", lua.LString(node.Version))
		L.RawSetInt(resultTable, i+1, nodeTable)
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetNamespaces(L *lua.LState) int {
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	namespaces, err := k8s.FetchNamespaces(client)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch namespaces: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, ns := range namespaces {
		L.RawSetInt(resultTable, i+1, lua.LString(ns))
	}
	L.Push(resultTable)
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetServiceAccounts(L *lua.LState) int {
	namespace := L.CheckString(1)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	serviceaccounts, err := k8s.GetServiceAccountsTableData(client, namespace)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch serviceaccounts: %v", err)))
		return 1
	}
	resultTable := L.NewTable()
	for i, sa := range serviceaccounts {
		saTable := L.NewTable()
		L.SetField(saTable, "Name", lua.LString(sa.Name))
		L.SetField(saTable, "Namespace", lua.LString(sa.Namespace))
		L.SetField(saTable, "Secrets", lua.LString(sa.Secrets))
		L.SetField(saTable, "Age", lua.LString(sa.Age))
		L.RawSetInt(resultTable, i+1, saTable)
	}
	L.Push(resultTable)
	return 1
}

// Delete functions
func (p *PluginmanagerStyleLuaPlugin) luaDeletePod(L *lua.LState) int {
	namespace := L.CheckString(1)
	name := L.CheckString(2)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	err := k8s.DeletePod(client, namespace, name)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to delete pod: %v", err)))
		return 1
	}
	L.Push(lua.LString("ok"))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaDeleteService(L *lua.LState) int {
	namespace := L.CheckString(1)
	name := L.CheckString(2)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	err := k8s.DeleteService(client, namespace, name)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to delete service: %v", err)))
		return 1
	}
	L.Push(lua.LString("ok"))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaDeleteDeployment(L *lua.LState) int {
	namespace := L.CheckString(1)
	name := L.CheckString(2)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	err := k8s.DeleteDeployment(client, namespace, name)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to delete deployment: %v", err)))
		return 1
	}
	L.Push(lua.LString("ok"))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaDeleteConfigMap(L *lua.LState) int {
	namespace := L.CheckString(1)
	name := L.CheckString(2)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	err := k8s.DeleteConfigmap(client, namespace, name)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to delete configmap: %v", err)))
		return 1
	}
	L.Push(lua.LString("ok"))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaDeleteSecret(L *lua.LState) int {
	namespace := L.CheckString(1)
	name := L.CheckString(2)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	err := k8s.DeleteSecret(client, namespace, name)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to delete secret: %v", err)))
		return 1
	}
	L.Push(lua.LString("ok"))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaDeleteIngress(L *lua.LState) int {
	namespace := L.CheckString(1)
	name := L.CheckString(2)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	err := k8s.DeleteIngress(client, namespace, name)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to delete ingress: %v", err)))
		return 1
	}
	L.Push(lua.LString("ok"))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaDeleteJob(L *lua.LState) int {
	namespace := L.CheckString(1)
	name := L.CheckString(2)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	err := k8s.DeleteJob(client, namespace, name)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to delete job: %v", err)))
		return 1
	}
	L.Push(lua.LString("ok"))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaDeleteCronJob(L *lua.LState) int {
	namespace := L.CheckString(1)
	name := L.CheckString(2)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	err := k8s.DeleteCronJob(client, namespace, name)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to delete cronjob: %v", err)))
		return 1
	}
	L.Push(lua.LString("ok"))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaDeleteDaemonSet(L *lua.LState) int {
	namespace := L.CheckString(1)
	name := L.CheckString(2)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	err := k8s.DeleteDaemonSet(client, namespace, name)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to delete daemonset: %v", err)))
		return 1
	}
	L.Push(lua.LString("ok"))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaDeleteStatefulSet(L *lua.LState) int {
	namespace := L.CheckString(1)
	name := L.CheckString(2)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	err := k8s.DeleteStatefulSet(client, namespace, name)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to delete statefulset: %v", err)))
		return 1
	}
	L.Push(lua.LString("ok"))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaDeleteReplicaSet(L *lua.LState) int {
	namespace := L.CheckString(1)
	name := L.CheckString(2)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	err := k8s.DeleteReplicaSet(client, namespace, name)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to delete replicaset: %v", err)))
		return 1
	}
	L.Push(lua.LString("ok"))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaDeleteServiceAccount(L *lua.LState) int {
	namespace := L.CheckString(1)
	name := L.CheckString(2)
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}
	err := k8s.DeleteServiceAccount(client, namespace, name)
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to delete serviceaccount: %v", err)))
		return 1
	}
	L.Push(lua.LString("ok"))
	return 1
}

func (p *PluginmanagerStyleLuaPlugin) luaGetEndpoints(L *lua.LState) int {
	namespace := L.CheckString(1)

	// Get the current client from the API
	client := p.api.GetClient()
	if client.Clientset == nil {
		L.Push(lua.LString("no kubernetes client available"))
		return 1
	}

	// Fetch endpoint slices data using the modern discovery.k8s.io/v1 API
	endpointSlices, err := client.Clientset.DiscoveryV1().EndpointSlices(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		L.Push(lua.LString(fmt.Sprintf("failed to fetch endpoint slices: %v", err)))
		return 1
	}

	// Convert to Lua table
	resultTable := L.NewTable()
	for i, endpointSlice := range endpointSlices.Items {
		// Format addresses
		var addresses []string
		for _, endpoint := range endpointSlice.Endpoints {
			for _, addr := range endpoint.Addresses {
				if addr != "" {
					// Check if endpoint is ready
					isReady := true
					if endpoint.Conditions.Ready != nil {
						isReady = *endpoint.Conditions.Ready
					}
					if !isReady {
						addresses = append(addresses, addr+" (not ready)")
					} else {
						addresses = append(addresses, addr)
					}
				}
			}
		}
		addressesStr := strings.Join(addresses, ", ")
		if addressesStr == "" {
			addressesStr = "<none>"
		}

		// Format ports
		var ports []string
		for _, port := range endpointSlice.Ports {
			if port.Port != nil {
				portStr := fmt.Sprintf("%d/%s", *port.Port, string(*port.Protocol))
				if port.Name != nil && *port.Name != "" {
					portStr = *port.Name + ":" + portStr
				}
				ports = append(ports, portStr)
			}
		}
		portsStr := strings.Join(ports, ", ")
		if portsStr == "" {
			portsStr = "<none>"
		}

		// Get associated service name from labels
		serviceName := "unknown"
		if endpointSlice.Labels != nil {
			if svcName, ok := endpointSlice.Labels["kubernetes.io/service-name"]; ok {
				serviceName = svcName
			}
		}

		// Calculate age
		age := "Unknown"
		if !endpointSlice.CreationTimestamp.IsZero() {
			age = format.FormatAge(endpointSlice.CreationTimestamp.Time)
		}

		endpointTable := L.NewTable()
		L.SetField(endpointTable, "Name", lua.LString(endpointSlice.Name))
		L.SetField(endpointTable, "Namespace", lua.LString(endpointSlice.Namespace))
		L.SetField(endpointTable, "Addresses", lua.LString(addressesStr))
		L.SetField(endpointTable, "Ports", lua.LString(portsStr))
		L.SetField(endpointTable, "Service", lua.LString(serviceName))
		L.SetField(endpointTable, "Age", lua.LString(age))

		L.RawSetInt(resultTable, i+1, endpointTable)
	}

	L.Push(resultTable)
	return 1
}
