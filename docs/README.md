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

`POST: /api/enable?id={id}&enable={enable}`

required **query** param:
`id:`  
just put the value that [Create](#Create-a-Job) returned.

**enable is true**:   
server would judge whether the operating conditions are met.
if not ,this api would return error info.

**enable is false**:  
just remove this job from schedule if it exists.

### Update Job Settings:

Sometimes we need change the task's execution strategy or basic information.
We can use this api.

`PUT : /api/job?id={id}`

Body:

```
{
    "name":"this is a new name",
    "cron":"* * 8 * * 2022",
}

```

the body is just a dictionary that
all fields you want to update should be contained.

the body allows follow fields:

- `name`
- `execType` set execType: timing job for `0`  delay job for `1`
- `cron` set the cron expression.Effective for timing job
- `description`
- `execAt` effective for delay job,it's timeStamp
- `script`

### Delete a job

### Run job directly

### Get job info