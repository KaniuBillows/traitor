package schedule

import (
	"context"
	"traitor/job"
)

type StandaloneSchedule struct {
	timeWheel *timeWheel
}

func MakeStandalone() Schedule {
	_ = &StandaloneSchedule{
		timeWheel: makeTimeWheel(),
	}
	return nil
}

func (s *StandaloneSchedule) Start(ctx context.Context) {
	s.timeWheel.start()
}
func (s *StandaloneSchedule) Close() {
	s.timeWheel.stop()
}

func (s *StandaloneSchedule) AddJob(j job.Job) error {

	s.timeWheel.AddJob(resolveCron(j.CronExpression), j.JobId, j.Function)
	return nil
}

func (s *StandaloneSchedule) CancelJob(key string) {
	s.timeWheel.removeTask(key)
}
