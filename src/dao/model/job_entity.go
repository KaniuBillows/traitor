package model

type JobEntity struct {
	Name    string
	JobId   string
	Script  string
	Cron    string
	JobType int
	State   int
}

type ScriptEntity struct {
	JobId  string
	Script string
}
