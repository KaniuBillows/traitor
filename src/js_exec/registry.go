package js_exec

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"traitor/js_exec/debug_out"
)

func RegistryModule(moduleName string, loader require.ModuleLoader) {
	if moduleName == console.ModuleName {
		return
	}
	registry.RegisterNativeModule(moduleName, loader)
	debugRegistry.RegisterNativeModule(moduleName, loader)
}

var registry = require.NewRegistry()
var debugRegistry = require.NewRegistry()

func init() {
	registry.RegisterNativeModule(console.ModuleName, console.Require)
	debugRegistry.RegisterNativeModule(debug_out.ModuleName, debug_out.Require)
}
func LoadModules(vm *goja.Runtime) {
	registry.Enable(vm)
	console.Enable(vm)
}

func LoadModulesForDebugMode(vm *goja.Runtime) {
	debugRegistry.Enable(vm)
	debug_out.Enable(vm)
}
