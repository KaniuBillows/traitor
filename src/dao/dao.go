package dao

import (
	"traitor/dao/localdb"
	"traitor/dao/model"
	"traitor/dao/mongoStoreage"
)

type Dao interface {
	GetJobInfos() ([]model.JobEntity, error)
	GetRunnableJobs() ([]model.JobEntity, error)
	GetJobInfo(jobId string) (model.JobEntity, error)
	GetJobScript(jobId string) (model.ScriptEntity, error)
	AddJob(job model.JobEntity) (string, error)
	UpdateJob(jobId string, mp map[string]any) error
	RemoveJob(jobId string) error
}

func CreateMongoDao(uri string, cluster string) Dao {
	return mongoStoreage.CreateMongoDao(uri)
}
func CreateLocalDao() Dao {
	return localdb.CreateLocalDao()
}
