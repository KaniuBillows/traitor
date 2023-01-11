# Traitor

Traitor is a distributed Timing Task Service.

It based on Golang but the tasks are described with JavaScript.

Users could use Web API to manage Timing work or Delay job.
It supported Dynamically add scheduled tasks without
republishing the program.

Besides,we provide a WebSite that you can manage tasks with UI.

# Getting Started

## docker

```
docker pull kaniu141/traitor  
docker run -p 8080:8080 -d traitor:latest
```

## build with source code

you should have environment: Go 1.19

```
git  clone https://github.com/KaniuBillows/traitor.git  
cd traitor/src  
go build  
.\traitor #start the service.
```

you can also use docker-compose:

```
git  clone https://github.com/KaniuBillows/traitor.git
cd traitor
docker-compose up -d
```

## startup parameters

| Parameter | Usage                                                            | Default |
|-----------|------------------------------------------------------------------|---------|
| -m        | the server is running standalone or with cluster.<br/> std/multi | std     |
| -r        | redis connection string. **required** if  running cluster        | -       |
| -mg       | mongodb connection string. **required** if running cluster       | -       |
| -c        | cluster name.only effective for cluster mode.                    | -       |
| -ip       | bind ip address. default will accept all.                        | -       |
| -p        | bind port                                                        | 8080    |

# Web API

## Job Management

### Create a Job

**Notice**: this API is just creating a job. The Job would be **stop state**
and would not be scheduled.

 ```
 POST:  /api/job?type={type}
 ```

required **query** param:  
`type:"timing" or "delay"` not case sensitive

body example **timing task**:  
required body param:`cron`  
must be a valid cron expression.
it could contain seconds part.

 ```
 {
    "name":"Every Monday 8:00am" ,
    "description":"say good morning to your friends",
    "cron":"0 0 8 ? * MON",
    "script":"//....."
 }
 ```

body example **delay task**:

required param:`execAt`   
must be **timeStamp** and the minim delay is 5 seconds.

```
{
    "name":"say happy new year",
    "description":"say happy new year to my friends at 2023 1.1 0:00 am",
    "execAt":"1672502400", //use time Stamp.
    "script":"//....."
}
```

success return :

```
{
    "id":"...." // job id would be returned.it's an uuid
}
```

### Start/Stop a job

```
POST: /api/enable?id={id}&enable={enable}
```

required **query** param:
`id:`  
just put the value that [Create](#Create-a-Job) returned.

the body allows these fields as follows:

- `name`
- `cron` set the cron expression.Effective for timing job
- `description`
- `execAt` effective for delay job,it's timeStamp
- `script`

**enable is true**:   
server would judge whether the operating conditions are met.
if not ,this api would return error info.

**enable is false**:  
just remove this job from schedule if it exists.

### Update Job Settings:

Sometimes we need change the task's execution strategy or basic information.
We can use this api.

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

the body is just a dictionary that
all fields you want to update should be contained.

the body allows these fields as follows:

- `name`
- `execType` set execType: timing job for `0`  delay job for `1`
- `cron` set the cron expression.Effective for timing job
- `description`
- `execAt` effective for delay job,it's timeStamp
- `script`

### Delete a job

```
DELTE /api/job?id={id}
```

### Get job info

```
GET /api/job?id={id}
```

this would **not** return the **script**.

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

### Get Script

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

### Run job directly

add a job and make it runnable.

this API will check all required param,
include `execType`and `execAt` or `cron`

``` 
POST /api/run
```

body:  
delay example:

```
{
    "name":"say happy new year",
    "description":"say happy new year to my friends at 2023 1.1 0:00 am",
    "execType":"1"
    "execAt":"1672502400", //use time Stamp.
    "script":"//....."
}
```

timing example:

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

### Debug a job

**Notice:** this is a websocket API

if a job is lunched by this api,the `console.log()` would print all result into
the **webSocket** client.

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

traitor is written by Golang but the tasks are javaScripts.The reason we do this is we want to keep the dynamic,
the business logic is rapid changing.

For running the js job,we choose the [Goja](https://github.com/dop251/goja) as the js runtime.
Goja contains full ECMAScript 5.1 support.But it's just a runtime
almost no standard library.Here's a [way](#plugin-development) that you can build
your **js library** with Golang.

So far we only support the JS **string**. Multi-file js program needs file system support,this will be supported in
the **future**
We plan to do it through mongoDB.

This script will be executed sequentially from top to bottom. Of course,
you can also define functions freely, as long
as your code conforms to the ES5 specification.

```
console.log("hello world") 

SayHello()

function SayHello(){
    console.log("hello")
}
```

# Cluster

The carrying capacity and throughput of a single node are limited,
and it cannot handle large-scale task scheduling.
You can deploy multi nodes with a simple way.

Still remember the [startup parameters](#startup-parameters)?

we can set the running mode with `-m`.just set `-m multi`.
And we need a redis server with `-r redis://localhost:6379`and a mongoDB with `mongodb://example.com:27017`.They are
indispensable under the distributed deployment
of traitor.

Then you can start traitor it or add more nodes.

There is an example that deploy two nodes with docker-compose:

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

**This is just an example,In a real distributed deployment, this method will not be used, but this example is to let
you know the dependencies between traitor and other services.**

You should decide how to deploy according to your own situation.

Just wait a moment, we forget a param `-c`,this will specify a **cluster name**.
all the nodes with the same cluster name will **share the task load**.

If you have multiple clusters,you don't need to prepare a redis and mongoDB for each cluster,
you can just set different cluster name.

# Working principle

For standalone mode,we implement a built-in K-V memory database and use
AOF persistence scheme just like Redis.

But for cluster,the local storage obviously can't meet the needs.So We choose
mongoDB as the storage.You should be flexible to choose mongoDB
configuration to suit your scenario.

Traitor uses **time-wheel algorithm** for task scheduling.
In cluster, each node holds all tasks that could be running.When tick triggers,
all subsequent tasks will be executed.So how to balance the execution load of the nodes?

**Distributed locks** are a viable solution,
but it also brings problem:the **time** of each node must be the same,
otherwise the task will not be executed evenly among the nodes.

In the end, we decided to determine which node should execute the task by
hashing the task id .When a node start, it will connect to the redis and
submit the node id with an expiration time of 1s as **heartbeat** cyclically.
Active nodes' heartbeats will form a set of nodes.Each node will
request the node set and hashing the task id,if the hashing result equals current node,
the job will be executed by current node.

And you might notice that traitor cluster doesn't contain a master node.It means
every node could receive the [Web Request](#web-api) or [debug](#debug-a-job) a job.

When a node receive a request, for example the `Run` request.It stores it in the database first,
then adds it to its own schedule.
After that, other nodes in the cluster will be notified through the **Pub-Sub mode** of redis

So you can set up a load balancer,
such as `nginx`, to provide a unified API entry for the traitor cluster.

# Plugin Development

todo...
