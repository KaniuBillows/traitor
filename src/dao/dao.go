package dao

import (
	"sync"
	"traitor/config"
	"traitor/dao/model"
	"traitor/dao/mongoStoreage"
)

type Dao interface {
	GetJobInfos() ([]model.JobEntity, error)
	GetJobInfo(jobId string) (*model.JobEntity, error)
	GetJobScript(jobId string) (*model.ScriptEntity, error)
	AddJob(job model.JobEntity) error
	UpdateJob(job model.JobEntity) error
	EditJobScript(jobId string, script string) error
	EditJobFiles() error
}

var once sync.Once
var dao Dao

func GetDao() Dao {
	once.Do(func() {
		dao = CreateDaoWithConfig()
	})
	return dao
}
func CreateDaoWithConfig() Dao {
	if config.GetConfig("runningMode") == "MultiNode" {
		var uri = config.GetConfig("mongoUri")
		return mongoStoreage.CreateMongoDao(uri)
	} else {
		return nil
	}
}
