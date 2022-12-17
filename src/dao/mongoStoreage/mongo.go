package mongoStoreage

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"traitor/dao"
	"traitor/dao/model"
	"traitor/job"
)

const (
	databaseName = "traitor_mongo"
	jobInfos     = "job_infos"
)

type MongoDao struct {
	c *mongo.Client
}

func CreateMongoDao(uri string) dao.Dao {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	var res = &MongoDao{
		c: client,
	}
	return res
}
func (m *MongoDao) GetJobInfos() ([]model.JobEntity, error) {
	coll := m.c.Database(databaseName).Collection(jobInfos)
	filter := bson.M{"state": job.Runnable}
	opt := options.Find().SetProjection(bson.M{
		"jobId":   1,
		"cron":    1,
		"jobType": 1,
	})
	cursor, err := coll.Find(context.TODO(), filter, opt)
	if err != nil {
		return nil, err
	}
	var res []model.JobEntity
	err = cursor.All(context.TODO(), &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *MongoDao) GetJobInfo(jobId string) (*model.JobEntity, error) {
	coll := m.c.Database(databaseName).Collection(jobInfos)
	filter := bson.M{"jobId": jobId}
	opt := options.FindOne().SetProjection(bson.M{
		"jobId":   1,
		"cron":    1,
		"jobType": 1,
	})
	var res model.JobEntity
	err := coll.FindOne(context.TODO(), filter, opt).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, err
}

func (m *MongoDao) GetJobScript(jobId string) (*model.ScriptEntity, error) {
	coll := m.c.Database(databaseName).Collection(jobInfos)
	filter := bson.M{"jobId": jobId, "jobType": job.JavaScriptJob}
	opt := options.FindOne().SetProjection(bson.M{
		"jobId":  1,
		"script": 1,
	})
	var res model.ScriptEntity
	err := coll.FindOne(context.TODO(), filter, opt).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, err
}

func (m *MongoDao) AddJob(job model.JobEntity) error {
	if job.JobId == "" {
		job.JobId = uuid.New().String()
	}

	coll := m.c.Database(databaseName).Collection(jobInfos)
	_, err := coll.InsertOne(context.TODO(), job)
	if err != nil {
		return err
	}
	return nil
}

func (m *MongoDao) UpdateJob(job model.JobEntity) error {
	if job.JobId == "" {
		return errors.New("job id cannot be empty")
	}
	coll := m.c.Database(databaseName).Collection(jobInfos)
	filter := bson.M{"jobId": job.JobId}
	update := bson.M{
		"jobName": job.Name,
		"state":   job.State,
		"cron":    job.Cron,
	}

	res := coll.FindOneAndUpdate(context.TODO(), filter, update)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (m *MongoDao) EditJobScript(jobId string, script string) error {
	coll := m.c.Database(databaseName).Collection(jobInfos)
	filter := bson.M{"jobId": jobId}
	update := bson.M{
		"script": script,
	}
	res := coll.FindOneAndUpdate(context.TODO(), filter, update)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (m *MongoDao) EditJobFiles() error {
	panic("not implementation")
}
