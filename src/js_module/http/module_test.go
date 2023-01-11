package http

import (
	"fmt"
	executor2 "github.com/KaniuBillows/traitor-plugin"
	"testing"
	"traitor/js_module"
)

func TestGet(t *testing.T) {
	var executor = executor2.MakeExecutor()
	js_module.RegistryAsyncPlugin(GetModule())
	js_module.LoadModules(executor)

	const script = `
	var Http=require('http')
	
	Http.get("Http://suggest.taobao.com/sug?code=utf-8&q=商品关键字&callback=cb",()=>{
  		console.log('hello world')
	},()=>{
  		console.log("err")
	})
`
	_, err := executor.Vm.RunString(script)
	executor.Wait.Wait()

	if err != nil {
		fmt.Println(err)
	}
}
