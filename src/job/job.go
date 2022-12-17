package job

import (
	"fmt"
	"github.com/dop251/goja"
	"io"
	"traitor/dao"
	"traitor/dao/model"
	"traitor/js_exec"
	"traitor/js_exec/debug_out"
	"traitor/logger"
)

const (
	JavaScriptJob      = 0
	GolangJob          = 1
	MultiFileScriptJob = 2
)
const (
	Runnable = 0
	Stop     = 1
)

type Job struct {
	JobId          string // should be unique.
	Function       func()
	CronExpression string
	JobType        int
}

func CreateWithEntity(entity *model.JobEntity) *Job {
	switch entity.JobType {
	case JavaScriptJob:
		{
			return CreateJsJob(entity)
		}
	case MultiFileScriptJob:
		{
			return CreateMultiFileJob()
		}
	default:
		// unsupported job type.
		return nil
	}
}
func CreateGolangJob(key string, f func(), cron string) *Job {
	var job = Job{
		JobId:          key,
		Function:       f,
		CronExpression: cron,
		JobType:        GolangJob,
	}
	return &job
}

func CreateJsJob(j *model.JobEntity) *Job {
	t := Job{}
	t.Function = func() {
		vm := goja.New()        // the vm is not concurrent safe.
		js_exec.LoadModules(vm) // native modules support.
		d := dao.GetDao()
		var sc, err = d.GetJobScript(j.JobId)
		if err != nil {
			logger.Error(fmt.Sprintf("running task failed:%s download script error.", j.JobId))
		}
		_, err = vm.RunString(sc.Script) // running logic.
		if err != nil {
			err.Error()
		}
	}
	return &t
}

// CreateJsJobForDebug
// Create a js job for debug.in this way,the js would log the console info into the writer.
func CreateJsJobForDebug(j model.JobEntity, writer io.Writer) *Job {
	vm := goja.New()
	js_exec.LoadModulesForDebugMode(vm)
	debug_out.SetIoWriter(vm, writer) // this vm would use this writer.
	d := dao.GetDao()
	var sc, err = d.GetJobScript(j.JobId)
	if err != nil {
		logger.Error(fmt.Sprintf("running task failed:%s download script error.", j.JobId))
	}
	_, err = vm.RunString(sc.Script) // running logic.
	if err != nil {
		err.Error()
	}
	return nil
}

func CreateMultiFileJob() *Job {
	// todo: impl this job type
	t := Job{}
	return &t
}
