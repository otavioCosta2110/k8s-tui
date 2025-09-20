package plugins

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/otavioCosta2110/k8s-tui/pkg/format"
	"github.com/otavioCosta2110/k8s-tui/pkg/k8s"
	"github.com/otavioCosta2110/k8s-tui/pkg/logger"
	"github.com/otavioCosta2110/k8s-tui/pkg/types"
	"github.com/yuin/gopher-lua"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PluginManager struct {
	registry             *PluginRegistry
	pluginDir            string
	luaStates            map[string]*lua.LState
	api                  *PluginAPIImpl
	pluginmanagerPlugins []PluginmanagerStylePlugin
}

func NewPluginManager(pluginDir string) *PluginManager {
	api := NewPluginAPI()
	return &PluginManager{
		registry:             NewPluginRegistry(),
		pluginDir:            pluginDir,
		luaStates:            make(map[string]*lua.LState),
		api:                  api,
		pluginmanagerPlugins: make([]PluginmanagerStylePlugin, 0),
	}
}

func (pm *PluginManager) LoadPlugins() error {
	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Starting plugin loading from directory: %s", pm.pluginDir))

	if pm.pluginDir == "" {
		logger.Info("🔌 Plugin Manager: No plugin directory specified, skipping plugin loading")
		return nil
	}

	// Check if plugin directory exists
	if _, err := os.Stat(pm.pluginDir); os.IsNotExist(err) {
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: Plugin directory does not exist: %s", pm.pluginDir))
		return nil
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Scanning for Lua plugins in: %s", pm.pluginDir))

	// Find all .lua files in the plugin directory and subdirectories
	var files []string
	err := filepath.Walk(pm.pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Error(fmt.Sprintf("🔌 Plugin Manager: Error scanning directory %s: %v", path, err))
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".lua") {
			files = append(files, path)
			logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Found potential plugin file: %s", path))
		}
		return nil
	})
	if err != nil {
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Failed to scan plugin directory: %v", err))
		return fmt.Errorf("failed to scan plugin directory: %v", err)
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Found %d potential plugin files", len(files)))

	loadedCount := 0
	failedCount := 0

	for _, file := range files {
		pluginName := strings.TrimSuffix(filepath.Base(file), ".lua")
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: Attempting to load plugin: %s from %s", pluginName, file))

		if err := pm.loadLuaPlugin(file); err != nil {
			logger.Error(fmt.Sprintf("🔌 Plugin Manager: ❌ Failed to load plugin %s: %v", pluginName, err))
			failedCount++
			continue
		}

		logger.Info(fmt.Sprintf("🔌 Plugin Manager: ✅ Successfully loaded plugin: %s", pluginName))
		loadedCount++
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Plugin loading complete - %d loaded, %d failed", loadedCount, failedCount))

	return nil
}

func (pm *PluginManager) loadLuaPlugin(path string) error {
	pluginName := strings.TrimSuffix(filepath.Base(path), ".lua")

	logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Creating Lua state for plugin: %s", pluginName))
	L := lua.NewState()

	// Variables for Neovim-style detection
	var setupType, configType, commandsType, hooksType lua.LValueType
	var isNeovimStyle bool

	// Load the Lua script
	logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Loading Lua script: %s", path))
	if err := L.DoFile(path); err != nil {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Failed to execute Lua script %s: %v", path, err))
		return fmt.Errorf("failed to load Lua script: %v", err)
	}
	logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Successfully loaded Lua script: %s", path))

	// Debug: Check what functions are available
	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Available functions in %s:", pluginName))
	for _, funcName := range []string{"Name", "Version", "Description", "Initialize", "Setup", "Config", "Commands", "Hooks", "GetResourceTypes", "GetUIExtensions"} {
		funcType := L.GetGlobal(funcName).Type()
		if funcType == lua.LTFunction {
			logger.Info(fmt.Sprintf("🔌 Plugin Manager:   %s: FUNCTION", funcName))
		} else {
			logger.Info(fmt.Sprintf("🔌 Plugin Manager:   %s: %s", funcName, funcType))
		}
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Validating required functions for plugin: %s", pluginName))

	// Check if required functions exist
	if L.GetGlobal("Name").Type() != lua.LTFunction {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Plugin %s missing required Name() function", pluginName))
		return fmt.Errorf("Lua plugin must define a Name function")
	}
	if L.GetGlobal("Initialize").Type() != lua.LTFunction {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Plugin %s missing required Initialize() function", pluginName))
		return fmt.Errorf("Lua plugin must define an Initialize function")
	}

	// Check if this is a Neovim-style plugin
	setupType = L.GetGlobal("Setup").Type()
	configType = L.GetGlobal("Config").Type()
	commandsType = L.GetGlobal("Commands").Type()
	hooksType = L.GetGlobal("Hooks").Type()

	isNeovimStyle = setupType == lua.LTFunction ||
		configType == lua.LTFunction ||
		commandsType == lua.LTFunction ||
		hooksType == lua.LTFunction

	// For Neovim-style plugins, set up the k8s_tui API before initialization
	if isNeovimStyle {
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎯 Detected pluginmanager-style plugin: %s", pluginName))
		logger.Info("🔌 Plugin Manager: Setting up k8s_tui API for pluginmanager-style plugin")

		// Create API table
		apiTable := L.NewTable()

		// Add API functions
		L.SetField(apiTable, "get_namespace", L.NewFunction(func(L *lua.LState) int {
			namespace := pm.api.GetCurrentNamespace()
			L.Push(lua.LString(namespace))
			return 1
		}))
		L.SetField(apiTable, "set_status", L.NewFunction(func(L *lua.LState) int {
			message := L.CheckString(1)
			pm.api.SetStatusMessage(message)
			return 0
		}))
		L.SetField(apiTable, "add_header", L.NewFunction(func(L *lua.LState) int {
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
			pm.api.AddHeaderComponent(component)
			return 0
		}))
		L.SetField(apiTable, "register_command", L.NewFunction(func(L *lua.LState) int {
			name := L.CheckString(1)
			description := L.CheckString(2)
			command := PluginCommand{
				Name:        name,
				Description: description,
				Handler: func(args []string) (string, error) {
					return "Command executed from Lua", nil
				},
			}
			pm.api.RegisterCommand(command.Name, command.Description, command.Handler)
			return 0
		}))
		// Kubernetes resource API functions
		L.SetField(apiTable, "get_pods", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_services", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_deployments", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_configmaps", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_secrets", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_ingresses", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_jobs", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_cronjobs", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_daemonsets", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_statefulsets", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_replicasets", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_nodes", L.NewFunction(func(L *lua.LState) int {
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_namespaces", L.NewFunction(func(L *lua.LState) int {
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "get_serviceaccounts", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			client := pm.api.GetClient()
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
		}))

		// Delete functions
		L.SetField(apiTable, "delete_pod", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "delete_service", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "delete_deployment", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "delete_configmap", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "delete_secret", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "delete_ingress", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "delete_job", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "delete_cronjob", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "delete_daemonset", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "delete_statefulset", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "delete_replicaset", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
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
		}))

		L.SetField(apiTable, "delete_serviceaccount", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)
			name := L.CheckString(2)
			client := pm.api.GetClient()
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
		}))

		// Keep the existing get_endpoints function
		L.SetField(apiTable, "get_endpoints", L.NewFunction(func(L *lua.LState) int {
			namespace := L.CheckString(1)

			// Get the current client from the API
			client := pm.api.GetClient()
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
		}))

		// Set the API in the global environment
		L.SetGlobal("k8s_tui", apiTable)
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: k8s_tui API set up for plugin: %s", pluginName))
	}

	// Create a LuaPlugin wrapper
	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Creating plugin wrapper for: %s", pluginName))
	luaPlugin := &LuaPlugin{
		L:          L,
		pluginName: pluginName,
	}

	// Get plugin metadata for logging
	pluginDisplayName := luaPlugin.Name()
	pluginVersion := luaPlugin.Version()
	pluginDescription := luaPlugin.Description()

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Initializing plugin %s (%s v%s)", pluginDisplayName, pluginName, pluginVersion))

	// Initialize the plugin
	if err := luaPlugin.Initialize(); err != nil {
		L.Close()
		logger.Error(fmt.Sprintf("🔌 Plugin Manager: Plugin %s initialization failed: %v", pluginDisplayName, err))
		return fmt.Errorf("failed to initialize Lua plugin: %v", err)
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Plugin %s initialized successfully", pluginDisplayName))

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Function types for %s - Setup: %s, Config: %s, Commands: %s, Hooks: %s",
		pluginName, setupType, configType, commandsType, hooksType))

	if isNeovimStyle {
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎯 Detected pluginmanager-style plugin: %s", pluginDisplayName))

		// Create Neovim-style plugin wrapper
		pluginmanagerPlugin := NewPluginmanagerStyleLuaPlugin(L, pluginName, pm.api)

		// Setup the plugin with default config
		defaultConfig := pluginmanagerPlugin.Config()
		if err := pluginmanagerPlugin.Setup(defaultConfig); err != nil {
			logger.Error(fmt.Sprintf("🔌 Plugin Manager: Failed to setup Neovim-style plugin %s: %v", pluginDisplayName, err))
			L.Close()
			return fmt.Errorf("failed to setup Neovim-style plugin: %v", err)
		}

		// Register commands
		commands := pluginmanagerPlugin.Commands()
		for _, cmd := range commands {
			pm.api.RegisterCommand(cmd.Name, cmd.Description, cmd.Handler)
		}

		// Register hooks
		hooks := pluginmanagerPlugin.Hooks()
		for _, hook := range hooks {
			pm.api.RegisterEventHandler(PluginEvent(hook.Event), hook.Handler)
		}

		pm.pluginmanagerPlugins = append(pm.pluginmanagerPlugins, pluginmanagerPlugin)
		logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎯 Registered pluginmanager-style plugin: %s v%s", pluginDisplayName, pluginVersion))

		// Also register as legacy plugin for backward compatibility
		hasResourcePlugin := luaPlugin.hasResourcePlugin()
		hasUIPlugin := luaPlugin.hasUIPlugin()

		if hasResourcePlugin {
			pm.registry.RegisterResourcePlugin(luaPlugin)
			logger.Info(fmt.Sprintf("🔌 Plugin Manager: 📊 Also registered as legacy resource plugin: %s", pluginDisplayName))
		}

		if hasUIPlugin {
			pm.registry.RegisterUIPlugin(luaPlugin)
			logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎨 Also registered as legacy UI plugin: %s", pluginDisplayName))
		}
	} else {
		// Handle legacy plugins (deprecated)
		logger.Warn(fmt.Sprintf("🔌 Plugin Manager: ⚠️  Legacy plugin detected: %s - Consider migrating to pluginmanager-style", pluginDisplayName))

		// Check plugin capabilities
		hasResourcePlugin := luaPlugin.hasResourcePlugin()
		hasUIPlugin := luaPlugin.hasUIPlugin()

		logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Plugin %s capabilities - Resource: %t, UI: %t", pluginDisplayName, hasResourcePlugin, hasUIPlugin))

		// Register based on capabilities
		if hasResourcePlugin {
			pm.registry.RegisterResourcePlugin(luaPlugin)
			logger.Info(fmt.Sprintf("🔌 Plugin Manager: 📊 Registered legacy resource plugin: %s v%s - %s", pluginDisplayName, pluginVersion, pluginDescription))

			// Log resource types
			resourceTypes := luaPlugin.GetResourceTypes()
			for _, rt := range resourceTypes {
				logger.Info(fmt.Sprintf("🔌 Plugin Manager:   └─ Resource type: %s (%s)", rt.Name, rt.Type))
			}
		}

		if hasUIPlugin {
			pm.registry.RegisterUIPlugin(luaPlugin)
			logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎨 Registered legacy UI plugin: %s v%s - %s", pluginDisplayName, pluginVersion, pluginDescription))
		}
	}

	// Store the Lua state
	pm.luaStates[pluginName] = L

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: 🎉 Plugin %s loaded and registered successfully", pluginDisplayName))

	return nil
}

func (pm *PluginManager) GetRegistry() *PluginRegistry {
	return pm.registry
}

func (pm *PluginManager) GetAPI() *PluginAPIImpl {
	return pm.api
}

func (pm *PluginManager) GetPluginmanagerPlugins() []PluginmanagerStylePlugin {
	return pm.pluginmanagerPlugins
}

func (pm *PluginManager) TriggerEvent(event PluginEvent, data interface{}) {
	pm.api.TriggerEvent(event, data)
}

func (pm *PluginManager) GetCustomResourceData(client k8s.Client, resourceType string, namespace string) ([]types.ResourceData, error) {
	// Set the client on the API so plugins can access it
	pm.api.SetClient(client)

	for _, plugin := range pm.registry.resourcePlugins {
		for _, rt := range plugin.GetResourceTypes() {
			if rt.Type == resourceType {
				return plugin.GetResourceData(client, resourceType, namespace)
			}
		}
	}
	return nil, fmt.Errorf("custom resource type %s not found", resourceType)
}

func (pm *PluginManager) DeleteCustomResource(client k8s.Client, resourceType string, namespace string, name string) error {
	for _, plugin := range pm.registry.resourcePlugins {
		for _, rt := range plugin.GetResourceTypes() {
			if rt.Type == resourceType {
				return plugin.DeleteResource(client, resourceType, namespace, name)
			}
		}
	}
	return fmt.Errorf("custom resource type %s not found", resourceType)
}

func (pm *PluginManager) GetCustomResourceInfo(client k8s.Client, resourceType string, namespace string, name string) (*k8s.ResourceInfo, error) {
	for _, plugin := range pm.registry.resourcePlugins {
		for _, rt := range plugin.GetResourceTypes() {
			if rt.Type == resourceType {
				return plugin.GetResourceInfo(client, resourceType, namespace, name)
			}
		}
	}
	return nil, fmt.Errorf("custom resource type %s not found", resourceType)
}

func (pm *PluginManager) Shutdown() error {
	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Shutting down %d loaded plugins", len(pm.luaStates)))

	shutdownCount := 0
	errorCount := 0

	for name, L := range pm.luaStates {
		logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Shutting down plugin: %s", name))

		// Call Shutdown if defined
		if L.GetGlobal("Shutdown").Type() == lua.LTFunction {
			logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Calling Shutdown() for plugin: %s", name))
			if err := L.CallByParam(lua.P{
				Fn:      L.GetGlobal("Shutdown"),
				NRet:    1,
				Protect: true,
			}); err != nil {
				logger.Error(fmt.Sprintf("🔌 Plugin Manager: Error calling Shutdown() for plugin %s: %v", name, err))
				errorCount++
			} else {
				// Check for error return
				ret := L.Get(-1)
				L.Pop(1)
				if ret.Type() == lua.LTString {
					logger.Error(fmt.Sprintf("🔌 Plugin Manager: Plugin %s shutdown returned error: %s", name, ret.String()))
					errorCount++
				} else {
					logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Plugin %s shutdown completed successfully", name))
					shutdownCount++
				}
			}
		} else {
			logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Plugin %s has no Shutdown() function, skipping", name))
		}

		L.Close()
		logger.Debug(fmt.Sprintf("🔌 Plugin Manager: Lua state closed for plugin: %s", name))
	}

	logger.Info(fmt.Sprintf("🔌 Plugin Manager: Shutdown complete - %d plugins shut down, %d errors", shutdownCount, errorCount))

	return nil
}
