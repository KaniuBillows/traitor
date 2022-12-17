package storeage

import (
	"github.com/dop251/goja"
	"strconv"
	"traitor/db/protocol"
)

func (d *DB) jsDel(call goja.FunctionCall) goja.Value {
	args := make([]string, len(call.Arguments))
	for i := 0; i < len(call.Arguments); i++ {
		args[i] = call.Arguments[i].String()
	}
	val, ok := d.del(args...)
	if ok {
		return d.runtime.ToValue(val)
	}
	return d.runtime.ToValue(nil)
}

func (d *DB) del(keys ...string) (int64, bool) {
	cmd := encodeCmd("DEL", keys...)
	var reply = d.client.Send(cmd)

	r, ok := reply.(*protocol.IntReply)
	if ok {
		return r.Code, ok
	}
	return -1, false
}

type expireOption struct {
	ex *int64
	px *int64
	xx bool
	nx bool
}

// jsSet
// @Description: set key value into the db.
// @receiver d *DB
// @param call goja.FunctionCall
//
// key:string;
//
// value:string;
//
// [options]:
//
//	{
//		ex:number,
//		px:number,
//		xx:boolean,
//		nx:boolean
//	}
//
// @return goja.Value
func (d *DB) jsSet(call goja.FunctionCall) goja.Value {
	key := call.Arguments[0].String()
	val := call.Arguments[1].String()
	if len(call.Arguments) > 2 {
		opt := call.Arguments[2].ToObject(d.runtime)
		if argCheck(opt) == false {
			return d.runtime.ToValue(d.set(key, val))
		}
		setOpt := expireOption{}
		ex := opt.Get("ex")
		if argCheck(ex) {
			exVal := ex.ToInteger()
			setOpt.ex = &exVal
		}
		px := opt.Get("px")
		if argCheck(px) {
			pxVal := px.ToInteger()
			setOpt.px = &pxVal
		}
		xx := opt.Get("xx")
		if argCheck(xx) {
			setOpt.xx = xx.ToBoolean()
		}
		nx := opt.Get("nx")
		if argCheck(nx) {
			setOpt.nx = nx.ToBoolean()
		}

		return d.runtime.ToValue(d.set1(key, val, setOpt))
	}
	return d.runtime.ToValue(d.set(key, val))
}
func (d *DB) set(key string, val string) bool {
	cmd := encodeCmd("SET", key, val)
	rep := d.client.Send(cmd)
	_, ok := rep.(*protocol.OkReply)
	return ok
}
func (d *DB) set1(key string, val string, option expireOption) bool {
	cmd := encodeCmd("SET", key, val)
	if option.ex != nil {
		cmd = append(cmd, encodeCmd("EX", strconv.FormatInt(*option.ex, 10))...)
	}
	if option.px != nil && option.px == nil {
		cmd = append(cmd, encodeCmd("PX", strconv.FormatInt(*option.px, 10))...)
	}
	if option.nx {
		cmd = append(cmd, encodeCmd("NX")...)
	} else if option.xx && option.nx == false {
		cmd = append(cmd, encodeCmd("XX")...)
	}
	rep := d.client.Send(cmd)
	_, ok := rep.(*protocol.OkReply)
	return ok
}

//		jsGetEx
//		@Description:
//		@receiver d *DB
//		@param call goja.FunctionCall
//
//		key:string
//
//	 	[option]:
//		{
//			ex:number,
//			px:number
//		}
//		@return goja.Value
func (d *DB) jsGetEx(call goja.FunctionCall) goja.Value {
	key := call.Arguments[0].String()
	epOption := expireOption{}
	if call.Argument(1) != goja.Undefined() {
		opt := call.Arguments[1].ToObject(d.runtime)
		if opt.Get("ex") != goja.Undefined() {
			exVal := opt.Get("ex").ToInteger()
			epOption.ex = &exVal
		}
		if opt.Get("px") != goja.Undefined() {
			pxVal := opt.Get("px").ToInteger()
			epOption.px = &pxVal
		}
	}
	return d.runtime.ToValue(d.getEx(key, epOption))
}
func (d *DB) getEx(key string, option expireOption) string {
	cmd := encodeCmd("GETEX", key)
	if option.ex != nil {
		cmd = append(cmd, encodeCmd("EX", strconv.FormatInt(*option.ex, 10))...)
	} else if option.px != nil && option.ex == nil {
		cmd = append(cmd, encodeCmd("PX", strconv.FormatInt(*option.px, 10))...)
	}
	rep := d.client.Send(cmd)
	switch rep.(type) {
	case *protocol.NullBulkReply:
		return ""
	case *protocol.BulkReply:
		res := rep.(*protocol.BulkReply)
		return decodeBytes(res.Arg)
	}
	return ""
}
