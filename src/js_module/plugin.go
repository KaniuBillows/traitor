//go:build linux

package js_module

import (
	"fmt"
	executor "github.com/KaniuBillows/traitor-plugin"
	"github.com/mitchellh/go-homedir"
	"os"
	"path"
	"plugin"
	"traitor/logger"
)

func init() {
	dirInit()
	watchDir()
}

var dir string

func dirInit() {
	d, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	d = fmt.Sprintf("%s/.traitor/plugin", d)
	err = os.MkdirAll(d, 0744)
	if err != nil {
		panic(err)
	}
	dir = d
}
func watchDir() {
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if path.Ext(name) != ".so" {
			continue
		}

		p, err := plugin.Open(dir + "/" + name)
		if err != nil {
			logger.Error(fmt.Sprintf("cannot load plugin file:%s  error:%s", name, err.Error()))
			continue
		}
		loadPlugin(name, p)
	}
}

func loadPlugin(moduleName string, p *plugin.Plugin) {
	moduleSymbol, err := p.Lookup("GetModule")
	if err != nil {
		logger.Error(err)
		return
	}
	switch moduleSymbol.(type) {
	case func() executor.Executable:
		{
			module := moduleSymbol.(func() executor.Executable)
			RegistryPlugin(module())
		}
	case func() executor.AsyncExecutable:
		{
			module := moduleSymbol.(func() executor.AsyncExecutable)
			RegistryAsyncPlugin(module())
		}
	default:
		logger.Error(fmt.Sprintf("load module file error: %s,Module not implemation AsyncExecutable or Executable", moduleName))
		return
	}
}
