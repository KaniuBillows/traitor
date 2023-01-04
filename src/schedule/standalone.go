package schedule

import (
	"context"
	"traitor/dao"
	"traitor/dao/model"
	"traitor/logger"
)

type StandaloneSchedule struct {
	schedule
}

func makeStandalone(d dao.Dao) *StandaloneSchedule {
	s := &StandaloneSchedule{
		schedule: schedule{timeWheel: makeTimeWheel(), dao: d},
	}
	return s
}

func (s *StandaloneSchedule) Start(_ context.Context) {
	s.timeWheel.start()
}

// load jobs from db.
func (s *StandaloneSchedule) initJobs() {
	jbs, err := s.dao.GetRunnableJobs()
	if err != nil {
		panic(err)
	}
	for _, jb := range jbs {
		_ = s.addJob(&jb)
	}
}
func (s *StandaloneSchedule) Close() {
	s.timeWheel.stop()
}

func (s *StandaloneSchedule) HandleJobStateChange(key string, state uint8) {
	if state == model.Runnable {
		j, err := s.dao.GetJobInfo(key)
		if err != nil {
			logger.Error(err)
			return
		}
		err = s.addJob(&j)
		if err != nil {
			logger.Error(err)
			return
		}
	} else {
		err := s.cancelJob(key)
		if err != nil {
			logger.Error(err)
			return
		}
	}
}
func (s *StandaloneSchedule) Remove(key string) {
	_ = s.cancelJob(key)
}

func (s *StandaloneSchedule) cancelJob(key string) error {
	s.timeWheel.removeJob(key)
	return nil
}

func (s *StandaloneSchedule) HandleJobTimeChange(key string) {
	jb, err := s.dao.GetJobInfo(key)
	if err != nil {
		logger.Error(err)
		return
	}
	err = s.addJob(&jb)
	if err != nil {
		logger.Error(err)
	}

}
