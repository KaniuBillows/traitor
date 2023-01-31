# Traitor

Traitor 是一个分布式定时任务服务

基于Golang构建，以JavaScript作为业务逻辑承载。

用户可以通过Web API与服务进行交互，管理定时或延时任务。它并非框架，支持动态添加任务
而无需任何重新发布。

同时，提供一个简易的Web端UI界面，可视化管理任务。

# 开始使用

## 基于Docker

```
docker pull kaniu141/traitor  
docker run -p 8080:8080 -d traitor:latest
```

## 基于源码构建

Golang环境: 1.19

```
git  clone https://github.com/KaniuBillows/traitor.git  
cd traitor/src  
go build  
```

或者使用docker-compose:

```
git  clone https://github.com/KaniuBillows/traitor.git
cd traitor
docker-compose up -d
```

## 启动参数

| 参数  | 用途                                     | 默认值  |
|-----|----------------------------------------|------|
| -m  | 以独立模式启动或者以集群模式启动.<br/> 值: std 或者 multi | std  |
| -r  | redis的连接字符串. **必须** 如果以集群模式启动          | -    |
| -mg | mongoDB连接字符串. **必须** 如果以集群模式启动         | -    |
| -c  | 集群名称，仅在集群模式生效                          | -    |
| -ip | 绑定ip地址，默认接受所有的来源                       | -    |
| -p  | 绑定的端口，默认                               | 8080 |

# Web API

## 任务管理

### 创建任务

**Notice**: 这个API仅仅会创建任务，任务会处于stop状态，并不会立即执行。

 ```
 POST:  /api/job?type={type}
 ```

必须参数 **query** :  
`type:"timing" or "delay"` 不区分大小写

body参数允许:

- `name`
- `cron` cron表达式，仅对定时任务生效
- `description`
- `execAt` 执行时间，仅对延时任务生效
- `script`

**定时任务**body实例:  
必须的body参数:`cron`  
必须为一个合法的cron表达式，它可以包含秒部分

 ```
 {
    "name":"Every Monday 8:00am" ,
    "description":"say good morning to your friends",
    "cron":"0 0 8 ? * MON",
    "script":"//....."
 }
 ```

**延时任务**body实例 :

必须参数:`execAt`   
必须为**时间戳**，且最小延时为5s

```
{
    "name":"say happy new year",
    "description":"say happy new year to my friends at 2023 1.1 0:00 am",
    "execAt":"1672502400", //use time Stamp.
    "script":"//....."
}
```

成功的返回值:

```
{
    "id":"...." // job id would be returned.it's an uuid
}
```

### 开始/停止一个任务

```
POST: /api/enable?id={id}&enable={enable}
```

必须的**query**参数:
`id:`  
只需要使用 [Create](#创建任务) 返回的id即可.

**enable为true**:   
服务将会判断任务是否满足启动条件，如果不满足，错误信息将会返回。

**enable为false**:  
直接将任务从调度器中移除。

### 更新任务设置:

有时我们需要更改任务的执行策略或基本信息。 我们可以使用这个api。

```
PUT : /api/job?id={id}
```

Body:

```
{
    "name":"this is a new name",
    "cron":"* * 8 * * 2022",
}

```

Body只是一个对象，应该包含您要更新的所有字段，正文允许这些字段如下：

- `name`
- `execType` set execType: timing job for `0`  delay job for `1`
- `cron` set the cron expression.Effective for timing job
- `description`
- `execAt` effective for delay job,it's timeStamp
- `script`

### 删除任务

```
DELTE /api/job?id={id}
```

### 获取任务信息

```
GET /api/job?id={id}
```

它将**不会**返回**脚本**.

return:

```
{
    "data":
    {
        "name":"Every Monday 8:00am" ,
        "description":"say good morning to your friends",
        "lastExecTime":"2022-12-04T07:17:58.782261202Z",
        "cron":"0 0 8 ? * MON"        
    }
}
```

### 获取脚本

```
GET /api/script?id={id}
```

return:

```
{
    "data":
    {
        "jobId":"xxx",
        "script":"xxx..."
    }
}
```

### 直接启动一个任务

添加一个任务，并尝试直接启动它

这个API会检查所有的参数，以确保能正常运行,
包含 `execType`以及 `execAt` 或 `cron`

``` 
POST /api/run
```

body:  
延时示例:

```
{
    "name":"say happy new year",
    "description":"say happy new year to my friends at 2023 1.1 0:00 am",
    "execType":"1"
    "execAt":"1672502400", //use time Stamp.
    "script":"//....."
}
```

定时示例:

```
{
    "name":"Every Monday 8:00am" ,
    "description":"say good morning to your friends",
    "cron":"0 0 8 ? * MON",
    "script":"//....."
 }
```

return

```
{
    "data":"jobid"
}
```

### 调试一个任务

**注意:** 这是一个websocket API

如果某个任务以这个API被激活,那么 `console.log()` 函数打印的所有信息，将被输出到
**webSocket** 客户端上.

url:

```
ws://{host}/api/debug
```

usage example:

script content:

```
console.log("hello world")
```

Client:

```
 let id = getId() // jobId  
 let host = window.location.host
 
 const ws = new WebSocket(`ws://${host}/api/debug?id=${id}`) // create a webSocket
 
 ws.addEventListener('message', e => {
        console.log(e.data) //the debug out result.
 })
    
```

result:

```
hello world
```

###

# JavaScript

traitor 是用 Golang 写的，但是任务是基于javaScript的。我们这样做的原因是我们想保留动态特性，
因为业务逻辑是快速变化的，如果纯使用Golang，那么程序的频繁发布将会是一个问题。

关于如何运行一个任务,我们选择了 [Goja](https://github.com/dop251/goja) 作为执行引擎.
Goja 支持完整的ECMAScript 5.1 特性.但是它却仅支持最基础的功能.但是有办法[way](#plugin-development)
可以让你通过golang，为脚本提供更丰富的功能支持。

目前我们仅支持**JS字符串**. 多文件任务需要文件系统的支持，但对于分布式系统来说也是一个挑战。在未来我们会考虑。

该脚本将从上到下依次执行。 当然你也可以自由定义函数，只要您的代码符合 ES5 规范。

```
console.log("hello world") 

SayHello()

function SayHello(){
    console.log("hello")
}
```

# 集群部署

单个节点的承载能力和吞吐量是有限的， 并且不能处理大规模的任务调度。
您可以通过简单的方式部署多个节点。

回顾一下 [startup parameters](#启动参数)

我们可以使用`-m`设置运行模式。只需设置`-m multi`。
我们需要一个redis并通过 `-r redis://localhost:6379`设置它，
以及一个mongoDB，通过 `mongodb://example.com:27017` 设置好它。
它们是分布式部署下不可或缺的。

然后你可以启动traitor并添加更多节点。

这是一个使用 docker-compose 部署两个节点的示例：

```
version: '3'

services:
    redis:
      image: "redis:7.0.7"  
      container_name: "redis"
      ports:
          - "6379:6379"
          
    mongo:
      image: "mongo:4.2.23"
      container_name: "mongo"
      ports:
          - "27017:27017"
                          
    node0:           
      image: "kaniu141/traitor:latest"
      ports:
          - "8080:8080"     
      volumes:
          - "~/.traitor/node0:/root/.traitor"
      container_name: "node0"
      depends_on:
          - redis
          - mongo
      command: -m multi -r redis://redis:6379 -mg mongodb://mongo:27017

              
    node1:
      image: "kaniu141/traitor:latest"
      ports:
          - "8081:8080"     
      volumes:
          - "~/.traitor/node1:/root/.traitor"
      container_name: "node1" 
      depends_on:
          - redis
          - mongo
      command: -m multi -r redis://redis:6379 -mg mongodb://mongo:27017     
```

**这只是一个例子，在真正的分布式部署中，不会用到这个方法，但是这个例子是为了让你知道traitor和其他服务之间的依赖关系。**

稍等一下，我们忘记了一个参数 `-c`，这将指定一个**集群名称**。
具有相同集群名称的所有节点将**共享任务负载**。

如果你有多个集群，你不需要为每个集群准备一个redis和mongoDB，你可以设置不同的集群名称。

# 工作原理

对于standalone模式，我们实现了一个内置的K-V内存数据库，和Redis一样使用AOF持久化方案。

但是对于集群来说，本地存储显然不能满足需求。所以我们选择mongoDB作为存储。你应该灵活地选择适合你场景的mongoDB配置。

Traitor使用**时间轮算法**进行任务调度。在集群中，每个节点都持有所有允许被运行的任务。
当tick触发时，所有后续任务都会被执行。那么如何平衡节点的执行负载呢？

**分布式锁**是一个可行的方案，
但也带来了问题：每个节点的**时间**必须相同，否则任务不会在节点间平均执行。

最后，我们决定通过一致性哈希来确定哪个节点应该执行任务。当一个节点启动时，
它会连接到redis并循环提交过期时间为1s的节点id作为**心跳**。
**活跃节点**的心跳会形成一个集合作为**一致性哈希环**。
每个节点都会请求哈希环并对任务id进行哈希处理，如果结果等于当前节点，则作业将由当前节点执行.

你可能会注意到叛徒集群不包含主节点。这意味着每个节点都可以接收 [Web Request](#web-api) 或者 [debug](#调试一个任务)请求.

当节点收到请求时，例如“运行”请求。它首先将其存储在数据库中，然后将其添加到自己的调度中。
之后，集群中的其他节点会通过redis的**Pub-Sub模式**得到通知，然后将任务状态同步到其他节点。

如果有节点下线，或者redis连接有问题，
**活动节点集合**将不包含该问题节点。因此不会有任何任务负载分配给有问题的节点。

从问题节点的角度来看，节点请求哈希环，但是得到一个错误，所以哈希环只是空的。
时间轮还在滴答作响，但是当节点尝试执行任务时，一致性hash的结果将永远不允许这样做，因为它是空的。

如果该节点能够恢复在线，心跳也会恢复，
集群中的其他节点也能感知到。任务负载分配恢复工作，每个节点将判断该任务是否应该由自己执行。

如果有新节点加入集群，过程也类似。

所以你不需要担心应该将请求发送到哪个节点。 可以搭建一个负载均衡器，比如`nginx`，
为traitor集群提供统一的API入口，因为集群中的节点是不区分master和slave的。

# 插件开发

traitor作为一个定时任务服务，不应该关心业务代码（指所有非定时相关的代码）。
但是矛盾的是：仅仅简单的javascript代码是不能满足需求的。
所以我们提供**plugin**，你可以用Golang编写自己的业务组件，通过**javaScript**来使用。

这是一个简单的例子，我们可以获得一个 http 客户端。

创建插件的mod:

```
mkdir http

cd http

go mod init http

#get the traitor-plugin
go get github.com/KaniuBillows/traitor-plugin 
``` 

代码部分：

```
package main

import (
	"C"
	executor "github.com/KaniuBillows/traitor-plugin"
	"github.com/dop251/goja"
	"io"
	"log"
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

func GetModule() executor.Executable {
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
			_, err := callback(goja.Undefined(), u.e.Vm.ToValue(rsp))
			if err != nil {
				log.Fatalln(err.Error())
			}
		}
	case err := <-errRc:
		{
			_, _ = errCallBack(goja.Undefined(), u.e.Vm.ToValue(err))
		}
	}
	u.e.Wait.Done()
}

```

构建插件:

```
go build -buildmode=plugin

cp http.so ~/.traitor/plugin/
```

再启动traitor，插件将会被加载.

这是可以通过js进行调用了:

```
var http=require('http')
	
	http.get("https://suggest.taobao.com/sug?code=utf-8&q=CS&callback=cb",()=>{
  		console.log('hello world')
	},()=>{
  		console.log("err")
	})
```

它将会输出 `heelo world` .

在这个例子中，你会注意到 `http.get()` 方法是一个异步函数，它不会阻塞执行，运行时将等待所有异步任务完成。

回到插件代码：
在`traitor-plugin`包中，有两个接口：`Executable`和`AsyncExecutable`。http插件需要异步执行，vm会通过`sync.WaitGroup`同步等待状态。

但是简单的同步插件要容易得多，只需 impl `Executable` 即可。

`GetName()` 方法声明如何获取插件，它的返回值将在 javaScript 中用作 `var plugin = require('pluginName')`。

`ModuleLoader` 用于设置运行时调用哪个native方法
并设置js函数名。   
对于这段代码：

```
func (m *Module) ModuleLoader(runtime *goja.Runtime, module *goja.Object) {
    m.vm = runtime
    obj := module.Get("exports").(*goja.Object)
    obj.Set("get", m.jsGet)
    }
}
```

javaScript函数名为`get`，调用时会执行golang方法`m.jsGet`。

Plugin是基于Golang的语言支持，只支持Linux,FreeBsd,macOS。