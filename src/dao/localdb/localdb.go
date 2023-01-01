package localdb

import (
	"context"
	"errors"
	"github.com/fatih/structs"
	"github.com/google/uuid"
	"strconv"
	"sync"
	"time"
	"traitor/dao/model"
	"traitor/db/client"
	"traitor/db/protocol"
	"traitor/db/startup"
	utils "traitor/db/util"
)

type LocalDb struct {
	client *client.Client
}

var dbClient *client.Client
var once sync.Once = sync.Once{}

func createClient() *client.Client {
	once.Do(func() {
		var ctx = context.Background()

		cnnFac, _ := startup.Startup(ctx)
		dbClient = cnnFac()
	})
	return dbClient
}
func CreateLocalDao() *LocalDb {
	c := createClient()
	return &LocalDb{
		client: c,
	}
}

const job_keys_set = "job_keys_set"

func (l *LocalDb) GetJobInfos() ([]model.JobEntity, error) {
	args := utils.ToCmdLine("SMembers", job_keys_set)
	reply := l.client.Send(args)
	var keys *protocol.MultiBulkReply
	switch reply.(type) {
	case *protocol.EmptyMultiBulkReply:
		{
			return make([]model.JobEntity, 0), nil
		}
	case *protocol.MultiBulkReply:
		{
			keys = reply.(*protocol.MultiBulkReply)
		}
	default:
		panic("unknown error:localdb job_keys_set was changed")
	}

	var res = make([]model.JobEntity, len(keys.Args))
	for i, key := range keys.Args {
		entity, err := l.GetJobInfo(string(key))
		if err != nil {
			continue
		}
		res[i] = entity
	}
	return res, nil
}
func toMap(args [][]byte, keys ...string) (map[string]string, error) {
	var res = make(map[string]string)
	if len(keys) != len(args) {
		return nil, errors.New("convert to map error.num of keys doesn't match")
	}
	for i, arg := range args {

		res[keys[i]] = string(arg)
	}
	return res, nil
}

func (l *LocalDb) GetRunnableJobs() ([]model.JobEntity, error) {
	all, err := l.GetJobInfos()
	if err != nil {
		return nil, err
	}
	result := make([]model.JobEntity, len(all))
	i := 0
	for _, j := range all {
		if j.State != model.Runnable {
			continue
		}
		result[i] = j
		i++
	}
	return result, nil
}

func (l *LocalDb) GetJobInfo(jobId string) (model.JobEntity, error) {

	cmd := utils.ToCmdLine("HMGET", jobId, model.Name, model.Cron, model.LastExecTime, model.State, model.Description,
		model.ExecType, model.ExecAt)
	reply := l.client.Send(cmd)
	multiBulkReply, ok := reply.(*protocol.MultiBulkReply)
	if ok == false {
		return model.JobEntity{}, errors.New("jobId is not exists")
	}
	var mp, err = toMap(multiBulkReply.Args, model.Name, model.Cron, model.LastExecTime, model.State, model.Description,
		model.ExecType, model.ExecAt)
	if err != nil {
		return model.JobEntity{}, err
	}

	var entity = model.JobEntity{
		JobId:       jobId,
		Name:        mp[model.Name],
		Cron:        mp[model.Cron],
		Description: mp[model.Description],
	}
	if mp[model.LastExecTime] != "" {
		t, err := time.Parse("2017-08-30 16:40:41", mp[model.LastExecTime])
		if err != nil {
			entity.LastExecTime = &t
		}
	}
	if mp[model.ExecAt] != "" {
		t, err := time.Parse("2017-08-30 16:40:41", mp[model.ExecAt])
		if err != nil {
			entity.ExecAt = &t
		}
	}
	state, err := strconv.ParseUint(mp[model.State], 10, 8)
	if err == nil {
		entity.State = uint8(state)
	}
	exeType, err := strconv.ParseUint(mp[model.ExecType], 10, 8)
	if err == nil {
		entity.ExecType = uint8(exeType)
	}

	return entity, nil
}
func (l *LocalDb) GetJobScript(jobId string) (model.ScriptEntity, error) {
	cmd := utils.ToCmdLine("HGET", jobId, model.Script)
	reply := l.client.Send(cmd)
	scriptReply, ok := reply.(*protocol.BulkReply)
	if ok == false {
		return model.ScriptEntity{}, errors.New("get job script error,script is not exists")
	}
	result := model.ScriptEntity{
		JobId:  jobId,
		Script: string(scriptReply.Arg),
	}
	return result, nil
}
func (l *LocalDb) AddJob(job model.JobEntity) error {
	if job.JobId == "" {
		job.JobId = uuid.NewString()
	}
	mp := structs.Map(job)
	args := make([]string, len(mp)*2+2)
	args[0] = "HMSET"
	args[1] = job.JobId
	i := 2
	for k, v := range mp {
		var value string
		switch v.(type) {
		case string:
			value = v.(string)
		case uint8:
			value = strconv.FormatUint(uint64(v.(uint8)), 10)
		case uint64:
			value = strconv.FormatUint(v.(uint64), 10)
		case time.Time:
			value = v.(time.Time).String()
		default:
			continue // ignore
		}
		args[i] = k
		args[i+1] = value
		i += 2
	}
	cmd := utils.ToCmdLine(args...)
	reply := l.client.Send(cmd)
	if status, ok := reply.(*protocol.StatusReply); ok == false || status.IsOKReply() == false {
		return errors.New("add failed")
	}
	setArg := make([]string, 3)
	setArg[0] = "SADD"
	setArg[1] = job_keys_set
	setArg[2] = job.JobId
	setCmd := utils.ToCmdLine(setArg...)
	setReply := l.client.Send(setCmd)
	if intReply, ok := setReply.(*protocol.IntReply); ok == false || intReply.Code != 1 {
		return errors.New("add failed")
	}
	return nil
}
func (l *LocalDb) UpdateJob(jobId string, mp map[string]any) error {
	if jobId == "" {
		return errors.New("job id cannot be empty")
	}
	key := jobId
	delete(mp, model.JobId)
	args := make([]string, len(mp)*2+2)
	args[0] = "HMSET"
	args[1] = key
	i := 2
	for k, v := range mp {
		var value string
		switch v.(type) {
		case string:
			value = v.(string)
		case uint8:
			value = strconv.FormatUint(uint64(v.(uint8)), 10)
		case uint64:
			value = strconv.FormatUint(v.(uint64), 10)
		case time.Time:
			value = v.(time.Time).String()
		default:
			continue // ignore
		}
		args[i] = k
		args[i+1] = value
		i += 2
	}
	cmd := utils.ToCmdLine(args...)
	reply := l.client.Send(cmd)
	if status, ok := reply.(*protocol.StatusReply); ok == false || status.IsOKReply() == false {
		return errors.New("add failed")
	}
	return nil
}
func (l *LocalDb) EditJobScript(jobId string, script string) error {
	args := make([]string, 4)
	args[0] = "HSET"
	args[1] = jobId
	args[2] = model.Script
	args[3] = script
	cmd := utils.ToCmdLine(args...)
	reply := l.client.Send(cmd)
	if _, ok := reply.(*protocol.IntReply); ok == false {
		return errors.New("update failed")
	}
	return nil
}

func (l *LocalDb) RemoveJob(jobId string) error {
	args := make([]string, 3)
	args[0] = "SREM"
	args[1] = job_keys_set
	args[2] = jobId
	cmd := utils.ToCmdLine(args...)
	reply := l.client.Send(cmd)
	if intReply, ok := reply.(*protocol.IntReply); ok == false || intReply.Code != 1 {
		return errors.New("remove failed")
	}

	args = make([]string, 2)
	args[0] = "DEL"
	args[1] = jobId
	cmd = utils.ToCmdLine(args...)
	reply = l.client.Send(cmd)
	if intReply, ok := reply.(*protocol.IntReply); ok == false || intReply.Code != 1 {
		return errors.New("remove failed")
	}
	return nil
}
