# Traitor

Traitor is a distributed Timing Task Service.

It based on Golang but the tasks are described with JavaScript.

Users could use Web API to manage Timing work or Delay job. It supported dynamically add scheduled tasks without republishing the program.

# Document
[Document](https://kaniubillows.github.io/traitor/#/) 

[中文文档](https://kaniubillows.github.io/traitor/#/zh-cn/)

# Quick Start

## docker

```
docker pull kaniu141/traitor
docker run -p 8080:8080 -d traitor:latest
```

## build with source code

```
git  clone https://github.com/KaniuBillows/traitor.git
cd traitor/src  
go build  
.\traitor
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
