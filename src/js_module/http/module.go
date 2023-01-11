package http

import (
	"C"
	executor "github.com/KaniuBillows/traitor-plugin"
	"github.com/dop251/goja"
	"io"
	httpclient "net/http"
)

const ModuleName = "http"

type Http struct {
	e *executor.Executor
}
type response struct {
	Status       string // e.g. "200 OK"
	StatusCode   int
	ResponseText string
}
type Module struct {
}

func (m *Module) GetName() string {
	return ModuleName
}

func GetModule() executor.AsyncExecutable {
	return &Module{}
}

func (m *Module) Require(e *executor.Executor) func(runtime *goja.Runtime, module *goja.Object) {
	u := Http{
		e: e,
	}
	return func(runtime *goja.Runtime, module *goja.Object) {
		obj := module.Get("exports").(*goja.Object)
		obj.Set("get", u.jsGet)
	}
}

func Get(url string, successRc chan *httpclient.Response, errRc chan *error) {
	res, err := httpclient.Get(url)
	if err != nil {
		errRc <- &err
		return
	}
	successRc <- res
}
func (u *Http) jsGet(call goja.FunctionCall) goja.Value {
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

	u.e.Wait.Add(1)
	go Get(url, rc, errRc)

	go u.callBack(callBack, errCallBack, rc, errRc)
	return goja.Undefined()
}
func (u *Http) callBack(callback goja.Callable,
	errCallBack goja.Callable,
	rc chan *httpclient.Response, errRc chan *error) {
	select {
	case res := <-rc:
		{
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(res.Body)
			body, _ := io.ReadAll(res.Body)
			str := string(body)

			var rsp = response{
				ResponseText: str,
				StatusCode:   res.StatusCode,
				Status:       res.Status,
			}
			_, _ = callback(goja.Undefined(), u.e.Vm.ToValue(rsp))
		}
	case err := <-errRc:
		{
			_, _ = errCallBack(goja.Undefined(), u.e.Vm.ToValue(err))
		}
	}
	u.e.Wait.Done()
}
