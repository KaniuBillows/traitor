package debug_out

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/dop251/goja_nodejs/util"
	"io"
	"os"
)

const ModuleName = "debug_console"

type Console struct {
	runtime *goja.Runtime
	util    *goja.Object
	writer  io.Writer
}

var defaultWriter io.Writer = os.Stdout

func (c *Console) log(p func(string)) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if format, ok := goja.AssertFunction(c.util.Get("format")); ok {
			ret, err := format(c.util, call.Arguments...)
			if err != nil {
				panic(err)
			}

			p(ret.String())
		} else {
			panic(c.runtime.NewTypeError("util.format is not a function"))
		}

		return nil
	}
}

func (c *Console) Log(s string) {
	_, _ = c.writer.Write([]byte(s))

}
func (c *Console) Error(s string) {
	_, _ = c.writer.Write([]byte(s))

}
func (c *Console) Warn(s string) {
	_, _ = c.writer.Write([]byte(s))
}
func Require(runtime *goja.Runtime, module *goja.Object) {
	requireWithPrinter(defaultWriter)(runtime, module)
}

func requireWithPrinter(writer io.Writer) require.ModuleLoader {
	return func(runtime *goja.Runtime, module *goja.Object) {
		c := &Console{
			runtime: runtime,
			writer:  writer,
		}

		c.util = require.Require(runtime, util.ModuleName).(*goja.Object)

		o := module.Get("exports").(*goja.Object)
		o.Set("log", c.log(c.Log))
		o.Set("error", c.log(c.Error))
		o.Set("warn", c.log(c.Warn))
	}
}
func Enable(runtime *goja.Runtime) {
	runtime.Set("console", require.Require(runtime, ModuleName))
}
func SetIoWriter(runtime *goja.Runtime, writer io.Writer) {
	var s = runtime.Get("console").(*goja.Object)
	var c = &Console{
		runtime: runtime,
		writer:  writer,
	}
	s.Set("log", c.log(c.Log))
	s.Set("error", c.log(c.Error))
	s.Set("warn", c.log(c.Warn))
}
