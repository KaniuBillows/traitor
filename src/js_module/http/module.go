package http

import (
	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"io"
	"log"
	httpclient "net/http"
)

const ModuleName = "http"

type http struct {
	runtime *goja.Runtime
}
type Response struct {
	Status       string // e.g. "200 OK"
	StatusCode   int
	ResponseText string
}

func Get(url string, successRc chan *httpclient.Response, errRc chan *error) {
	res, err := httpclient.Get(url)
	if err != nil {
		errRc <- &err
		return
	}
	successRc <- res
}
func (u *http) jsGet(call goja.FunctionCall) goja.Value {
	var url string

	rc := make(chan *httpclient.Response)
	errRc := make(chan *error)
	if arg := call.Argument(0); !goja.IsUndefined(arg) {
		url = arg.String()
	}
	var callBack goja.Callable
	if _callBack, ok := goja.AssertFunction(call.Argument(1)); ok {
		callBack = _callBack
	}
	var errCallBack goja.Callable
	if _errCallBack, ok := goja.AssertFunction(call.Argument(2)); ok {
		errCallBack = _errCallBack
	}
	go Get(url, rc, errRc)

	go u.callBack(callBack, errCallBack, rc, errRc)
	return goja.Undefined()
}

func Require(runtime *goja.Runtime, module *goja.Object) {
	u := &http{
		runtime: runtime,
	}
	obj := module.Get("exports").(*goja.Object)
	obj.Set("get", u.jsGet)
}

func (u *http) callBack(callback goja.Callable,
	errCallBack goja.Callable,
	rc chan *httpclient.Response, errRc chan *error) {
	select {
	case res := <-rc:
		{
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)
			str := string(body)

			var rsp = Response{
				ResponseText: str,
				StatusCode:   res.StatusCode,
				Status:       res.Status,
			}
			_, err := callback(goja.Undefined(), u.runtime.ToValue(rsp))
			if err != nil {
				log.Fatalln(err.Error())
			}
		}
	case err := <-errRc:
		{
			_, _ = errCallBack(goja.Undefined(), u.runtime.ToValue(err))
		}
	}
}

func init() {
	require.RegisterNativeModule(ModuleName, Require)
}
