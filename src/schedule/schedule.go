package schedule

import (
	"context"
	"errors"
	"fmt"
	executor "github.com/KaniuBillows/traitor-plugin"
	"github.com/gorhill/cronexpr"
	"io"
	"sync"
	"time"
	"traitor/dao"
	"traitor/dao/model"
	"traitor/js_module"
	"traitor/js_module/debug_out"
	"traitor/logger"
)

// Schedule just exec the job and notify other nodes.
type Schedule interface {
	Close()
	Start(ctx context.Context)
	//HandleJobStateChange
	//handle enable or disable a job.
	HandleJobStateChange(key string, state uint8)
	//HandleJobTimeChange
	// handle cron or delay change.
	HandleJobTimeChange(key string)
	CreateTask(key string, execType uint8) func()
	CreateTaskForDebug(key string, writer io.Writer) (func(), *sync.WaitGroup)
	ResolveCron(str string) (time.Duration, error)
	Remove(key string)
}
type schedule struct {
	dao       dao.Dao
	timeWheel *timeWheel
}

func (s *schedule) CreateTask(key string, execType uint8) func() {

	execFunc := func() {
		exec := executor.MakeExecutor()
		js_module.LoadModules(exec) // native modules support.
		var sc, err = s.dao.GetJobScript(key)
		if err != nil {
			logger.Error(fmt.Sprintf("running Task failed:%s download script error.", key))
			return
		}
		_, err = exec.Vm.RunString(sc.Script) // running logic.
		if err != nil {
			logger.Error(err)
		}
		// update last exec time
		err = s.dao.UpdateJob(key, map[string]any{model.LastExecTime: time.Now()})
		if err != nil {
			logger.Error(err)
		}
		exec.Wait.Wait()
	}
	if execType == model.DelayExecute { // only once for delay.
		return execFunc
	} else {
		return func() {
			execFunc()
			// after execute re-add into for next time.
			j, err := s.dao.GetJobInfo(key)
			if err != nil {
				logger.Error(fmt.Sprintf("re-add timing job error, cannot get the job entity:%s", key))
				return
			}
			err = s.addJob(&j)
			if err != nil {
				logger.Error(fmt.Sprintf("re-add timing job error:  %s", err.Error()))
			}
		}
	}
}
func (s *schedule) addJob(j *model.JobEntity) error {
	fn := s.CreateTask(j.JobId, j.ExecType)
	if j.ExecType == model.TimingExecute {
		d, err := s.ResolveCron(j.Cron)
		if err != nil {
			return err
		}
		s.timeWheel.AddJob(d, j.JobId, fn)
	} else {
		delay := j.ExecAt.ToTime().Sub(time.Now())
		if delay <= 0 {
			return errors.New("delay job has expired")
		}
		s.timeWheel.AddJob(delay, j.JobId, fn)
	}
	return nil
}

func (s *schedule) CreateTaskForDebug(key string, writer io.Writer) (func(), *sync.WaitGroup) {
	wt := sync.WaitGroup{}
	wt.Add(1)
	return func() {
		defer func() {
			wt.Done()
		}()

		exec := executor.MakeExecutor()
		js_module.LoadModulesForDebugMode(exec)
		debug_out.SetIoWriter(exec.Vm, writer) // this vm would use this writer.
		var sc, err = s.dao.GetJobScript(key)
		if err != nil {
			logger.Error(fmt.Sprintf("running Task failed:%s download script error.", key))
			return
		}
		_, err = exec.Vm.RunString(sc.Script) // running logic.
		if err != nil {
			errInfo := err.Error()
			buffer := []byte(errInfo)
			_, err = writer.Write(buffer)
			if err != nil {
				logger.Error(err)
			}
		}
		exec.Wait.Wait()
	}, &wt
}

// ResolveCron
// return the delay time of the cron.
func (s *schedule) ResolveCron(str string) (time.Duration, error) {
	t := cronexpr.MustParse(str).Next(time.Now())
	if t.IsZero() == true {
		return time.Second * 0, errors.New("job would never get next exec time")
	}
	return t.Sub(time.Now()), nil
}
func StartMultiNode(redisStr string, mongoUri string, cluster string) (Schedule, dao.Dao) {
	d := dao.CreateMongoDao(mongoUri, cluster)
	schedule := makeMultiNode(redisStr, d, cluster)
	ctx := context.Background()
	schedule.Start(ctx)

	return schedule, d
}

func StartStandalone() (Schedule, dao.Dao) {
	d := dao.CreateLocalDao()
	schedule := makeStandalone(d)
	ctx := context.Background()
	schedule.Start(ctx)
	return schedule, d
}
