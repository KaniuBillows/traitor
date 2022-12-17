package schedule

import (
	"context"
	"time"
	"traitor/job"
)

type Schedule interface {
	Close()
	Start(ctx context.Context)
	AddJob(j job.Job) error
	CancelJob(key string, jobType int) error
	//UpdateTask()
}

const (
	Standalone = 0
	MultiNodes = 1
)

// return the delay time of the cron. todo: impl this method.
func resolveCron(str string) time.Duration {
	return 0
}
