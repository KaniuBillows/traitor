//go:build linux

package js_module

import (
	"fmt"
	"github.com/dop251/goja"
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
	dir, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	dir = fmt.Sprintf("%s/.traitor/plugin", dir)
	err = os.MkdirAll(dir, 0744)
	if err != nil {
		panic(err)
	}
}
func watchDir() {
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if path.Ext(f.Name()) != "so" {
			continue
		}

		p, err := plugin.Open(f.Name())
		if err != nil {
			logger.Error(fmt.Sprintf("cannot load plugin file:%s", f.Name()))
			continue
		}
		loadPlugin(f.Name(), p)
	}
}

func loadPlugin(moduleName string, p *plugin.Plugin) {
	nameSymbol, err := p.Lookup("ModuleName")
	if err != nil {
		logger.Error(err)
		return
	}
	name, ok := nameSymbol.(string)
	if ok == false {
		logger.Error(fmt.Sprintf("load module error: %s,ModuleName required", moduleName))
		return
	}
	fnSymbol, err := p.Lookup("Require")
	if err != nil {
		logger.Error(err)
		return
	}
	fn, ok := fnSymbol.(func(runtime *goja.Runtime, module *goja.Object))
	if ok == false {
		logger.Error(fmt.Sprintf("load module error: %s,func 'Require' is required", moduleName))
		return
	}

	RegistryModule(name, fn)
}
