package http

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"testing"
)

func TestGet(t *testing.T) {
	vm := goja.New()
	registry := require.NewRegistry()

	registry.RegisterNativeModule(ModuleName, Require)
	registry.RegisterNativeModule(console.ModuleName, console.Require)
	registry.Enable(vm)
	console.Enable(vm)
	//req := registry.Enable(vm)
	//v, _ := req.Require("node:console")
	const script = `
	//const http = require("node")
	//var result
	//http.get('https://api.kaniu.pro/ne/search?keyword=%E9%98%BF%E7%89%9B',(e)=>{
	//	var obj = JSON.parse(e.ResponseText)
	//	console.logger(JSON.stringify(obj[0]))
	//	result = obj
	//},(err)=>{})
	//var console = require("node:console")
	console.log("hello world")
`
	_, err := vm.RunString(script)
	//var l = v.(*goja.Object).Get("log")

	if err != nil {
		//fmt.Println(l)
	}
}
