package js_module

import (
	executor "github.com/KaniuBillows/traitor-plugin"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/dop251/goja_nodejs/util"
	"traitor/js_module/debug_out"
	"traitor/js_module/http"
)

// register inside modules.
func init() {
	RegistryAsyncPlugin(http.GetModule())
}

func RegistryPlugin(p executor.Executable) {
	// add into global registry
	require.RegisterNativeModule(p.GetName(), p.ModuleLoader)
}

func RegistryAsyncPlugin(p executor.AsyncExecutable) {
	plugins = append(plugins, p)
}

var plugins = make([]executor.AsyncExecutable, 0)

func LoadModules(exec *executor.Executor) {
	if len(plugins) > 0 {
		var registry = require.NewRegistry()
		for _, plugin := range plugins {
			loader := plugin.Require(exec)
			name := plugin.GetName()
			registry.RegisterNativeModule(name, loader)
		}
		registry.Enable(exec.Vm)
	}
	console.Enable(exec.Vm)
}

func LoadModulesForDebugMode(exec *executor.Executor) {
	var registry = require.NewRegistry()
	if len(plugins) > 0 {
		for _, plugin := range plugins {
			loader := plugin.Require(exec)
			name := plugin.GetName()
			registry.RegisterNativeModule(name, loader)
		}
		registry.Enable(exec.Vm)
	}
	registry.RegisterNativeModule(debug_out.ModuleName, debug_out.Require)
	registry.RegisterNativeModule(util.ModuleName, util.Require)
	registry.Enable(exec.Vm)
	debug_out.Enable(exec.Vm)
}
