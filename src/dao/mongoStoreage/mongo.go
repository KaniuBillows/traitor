package mongoStoreage

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"traitor/dao/model"
)

const (
	databaseName = "traitor_mongo"
	jobInfos     = "job_infos"
)

type MongoDao struct {
	c *mongo.Client
}

func CreateMongoDao(uri string) *MongoDao {
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
	opt := options.Find().SetProjection(bson.M{
		"jobId":   1,
		"cron":    1,
		"jobType": 1,
	})
	cursor, err := coll.Find(context.TODO(), bson.M{}, opt)
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
func (m *MongoDao) GetRunnableJobs() ([]model.JobEntity, error) {
	coll := m.c.Database(databaseName).Collection(jobInfos)
	filter := bson.M{"state": model.Runnable}
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
func (m *MongoDao) GetJobInfo(jobId string) (model.JobEntity, error) {
	coll := m.c.Database(databaseName).Collection(jobInfos)
	filter := bson.M{"jobId": jobId}
	opt := options.FindOne().SetProjection(bson.M{
		model.JobId:        1,
		model.Cron:         1,
		model.Name:         1,
		model.LastExecTime: 1,
		model.State:        1,
		model.ExecType:     1,
		model.Description:  1,
		model.ExecAt:       1,
	})
	var res model.JobEntity
	err := coll.FindOne(context.TODO(), filter, opt).Decode(&res)
	if err != nil {
		return res, err
	}
	return res, err
}

func (m *MongoDao) GetJobScript(jobId string) (model.ScriptEntity, error) {
	coll := m.c.Database(databaseName).Collection(jobInfos)
	filter := bson.M{"jobId": jobId}
	opt := options.FindOne().SetProjection(bson.M{
		"jobId":  1,
		"script": 1,
	})
	var res model.ScriptEntity
	err := coll.FindOne(context.TODO(), filter, opt).Decode(&res)
	if err != nil {
		return res, err
	}
	return res, err
}

func (m *MongoDao) AddJob(job model.JobEntity) (string, error) {
	if job.JobId == "" {
		job.JobId = uuid.New().String()
	}
	coll := m.c.Database(databaseName).Collection(jobInfos)
	_, err := coll.InsertOne(context.TODO(), job)
	if err != nil {
		return job.JobId, err
	}
	return job.JobId, nil
}

func (m *MongoDao) UpdateJob(jobId string, mp map[string]any) error {
	if jobId == "" {
		return errors.New("job id cannot be empty")
	}
	coll := m.c.Database(databaseName).Collection(jobInfos)
	filter := bson.M{model.JobId: jobId}
	delete(mp, model.JobId)

	res := coll.FindOneAndUpdate(context.TODO(), filter, mp)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (m *MongoDao) EditJobScript(jobId string, script string) error {
	coll := m.c.Database(databaseName).Collection(jobInfos)
	filter := bson.M{model.JobId: jobId}
	update := bson.M{
		model.Script: script,
	}
	res := coll.FindOneAndUpdate(context.TODO(), filter, update)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (m *MongoDao) RemoveJob(jobId string) error {
	coll := m.c.Database(databaseName).Collection(jobInfos)
	filter := bson.M{"jobId": jobId}
	_, err := coll.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	return nil
}

func (m *MongoDao) EditJobFiles() error {
	panic("not implementation")
}
