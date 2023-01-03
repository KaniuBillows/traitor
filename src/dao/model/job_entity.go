package model

import (
	"time"
)

const (
	TimingExecute = 0
	DelayExecute  = 1
)
const (
	Runnable = 1
	Stop     = 0
)

type JobEntity struct {
	Name         string     `json:"name,omitempty" bson:"name,omitempty"  structs:"name,omitempty"`
	JobId        string     `json:"jobId,omitempty" bson:"jobId,omitempty"  structs:"jobId,omitempty"`
	Cron         string     `json:"cron,omitempty" bson:"cron,omitempty"  structs:"cron,omitempty"`
	Description  string     `json:"description" bson:"description,omitempty" structs:"description,omitempty"`
	LastExecTime *time.Time `json:"lastExecTime,omitempty" bson:"lastExecTime,omitempty" structs:"lastExecTime,omitempty"`
	ExecAt       *TimeStamp `json:"execAt,omitempty" bson:"execAt,omitempty" structs:"execAt,omitempty"`
	ExecType     uint8      `json:"execType" bson:"execType" structs:"execType"`
	State        uint8      `json:"state" bson:"state" structs:"state"`
	Script       string     `json:"script" bson:"script" structs:"script,omitempty"`
}

const (
	JobId        = "jobId"
	Script       = "script"
	Name         = "name"
	LastExecTime = "lastExecTime"
	Description  = "description"
	Cron         = "cron"
	ExecType     = "execType"
	ExecAt       = "execAt"
	State        = "state"
)

type ScriptEntity struct {
	JobId  string `json:"jobId,omitempty" bson:"jobId,omitempty"`
	Script string `json:"script,omitempty" bson:"script,omitempty"`
}
