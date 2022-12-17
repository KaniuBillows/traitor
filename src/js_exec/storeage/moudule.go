package storeage

import (
	"context"
	"github.com/dop251/goja"
	"traitor/db/client"
	"traitor/db/startup"
)

const ModuleName = "db"

type DB struct {
	runtime *goja.Runtime
	client  *client.Client
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	var ctx = context.Background()

	cnnFac, _ := startup.Startup(ctx)
	db := &DB{
		runtime: runtime,
		client:  cnnFac(),
	}
	obj := module.Get("exports").(*goja.Object)
	_ = obj.Set("Del", db.jsDel)
	_ = obj.Set("Set", db.jsSet)
}

/***************Some Value Convert************************/
func encodeCmd(cmd string, ags ...string) [][]byte {
	args := make([][]byte, len(ags)+1)
	args[0] = []byte(cmd)
	for i, s := range ags {
		args[i+1] = []byte(s)
	}
	return args
}
func decodeBytes(bytes []byte) string {
	return string(bytes)
}

func argCheck(value goja.Value) bool {
	return value != goja.Undefined() && value != nil
}

/******************End***********************************/
