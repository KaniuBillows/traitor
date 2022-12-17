package schedule

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"strings"
	"time"
	"traitor/consistenthash"
	"traitor/dao"
	"traitor/job"
	"traitor/logger"
)

const (
	nodeId           = "traitor_defaultNode_" //todo: with config.
	heartBeatTimeout = 1                      //doesn't update node's status within 5s,indicating that this node is down. todo: with config.
	replicaCount     = 50
	redisChannel     = "traitor_pub_sub"
)

type MultiNodeSchedule struct {
	timeWheel     *timeWheel
	client        *redis.Client
	NodeId        string
	cancel        context.CancelFunc
	consistentMap *consistenthash.Map
	dao           dao.Dao
}

func MakeMultiNode(connectionStr string) Schedule {
	opt, err := redis.ParseURL(connectionStr)
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(opt)
	if err != nil {
		panic(err.Error())
	}
	uid := uuid.New()
	s := &MultiNodeSchedule{
		timeWheel: makeTimeWheel(),
		client:    client,
		NodeId:    nodeId + uid.String(),
	}
	return s
}
func (s *MultiNodeSchedule) Close() {
	s.timeWheel.stop()
	_ = s.client.Close()

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *MultiNodeSchedule) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	s.startHeartBeat(ctx)
	s.startSyncNodeList(ctx)
	s.startSub(ctx)
	s.timeWheel.start()
}
func (s *MultiNodeSchedule) syncNodeList(ctx context.Context) {
	var cursor uint64 = 0
	list := make([]string, 0)
	matchStr := fmt.Sprintf("%s*", nodeId)
	for {
		scan := s.client.Scan(ctx, cursor, matchStr, 1000)
		keys, cursor, err := scan.Result()
		if err != nil {
			logger.Error("cannot sync node list from redis")
			s.Close()
			return
		}
		list = append(list, keys...)
		if cursor == 0 {
			break
		}
	}
	s.consistentMap = consistenthash.New(replicaCount, nil)
	s.consistentMap.Add(list...)
}
func (s *MultiNodeSchedule) startSyncNodeList(ctx context.Context) {
	go func() {
		var c = time.Tick(time.Second * heartBeatTimeout)
		for {
			select {
			case <-c:
				{
					// pull the node list.
					s.syncNodeList(ctx)
				}
			case <-ctx.Done():
				{
					return
				}
			}
		}
	}()
}

func (s *MultiNodeSchedule) startHeartBeat(ctx context.Context) {
	go func() {
		var c = time.Tick(time.Second) // 1s update.
		for {
			select {
			case <-c:
				{
					res := s.client.SetEX(ctx, s.NodeId, "", time.Second*heartBeatTimeout)
					if res.Err() != nil {
						// sth wrong with the redis connection.
						logger.Error(res.Err())
						s.Close()
					}
				}
			case <-ctx.Done():
				{
					return
				}
			}
		}
	}()
}

func (s *MultiNodeSchedule) AddJob(j job.Job) error {

	// if js job
	// publish job id to redis for notify other nodes.
	// If the job has already been in the timeWheel, it should be re-added.
	if j.JobType == job.JavaScriptJob {
		s.timeWheel.AddJob(resolveCron(j.CronExpression), j.JobId, j.Function)
		//notify other nodes.
		res := s.client.Publish(context.TODO(), redisChannel, fmt.Sprintf(jobAdd, j.JobId))
		if res.Err() != nil {
			return res.Err()
		}
		return nil
	}

	// if it is go job:
	if j.JobType == job.GolangJob {
		delay := resolveCron(j.CronExpression)

		s.timeWheel.AddJob(delay*time.Second, j.JobId, j.Function)
	}
	return nil
}

// CancelJob remove the job from the time wheel.
func (s *MultiNodeSchedule) CancelJob(key string, jobType int) error {
	// if js job
	if jobType == job.JavaScriptJob {
		// notify other nodes.
		res := s.client.Publish(context.TODO(), redisChannel, fmt.Sprintf(jobCancel, key))
		if res.Err() != nil {
			return res.Err()
		}
		s.timeWheel.removeTask(key)
		return nil
	}
	// if golang job:
	if jobType == job.GolangJob {
		s.timeWheel.removeTask(key)
	}
	return nil
}

func (s *MultiNodeSchedule) startSub(ctx context.Context) {
	pubSub := s.client.Subscribe(ctx, redisChannel)
	subChanel := pubSub.Channel()
	go func() {
		select {
		case <-ctx.Done():
			{
				_ = pubSub.Close()
				return
			}
		case msg := <-subChanel: // get startSub msg.
			{
				s.handleSubEvent(msg)
			}
		}
	}()
}

func (s *MultiNodeSchedule) handleSubEvent(msg *redis.Message) {
	// job add.
	if strings.Contains(msg.Payload, "jobAdd") {
		res := strings.Split(msg.Payload, ":")
		id := res[1]
		entity, err := s.dao.GetJobInfo(id)
		if err != nil {
			logger.Error(fmt.Sprintf("could not load the job from db.id:%s", id))
		}
		task := job.CreateWithEntity(entity)
		s.timeWheel.AddJob(resolveCron(task.CronExpression), task.JobId, task.Function)
		return
	}

	// job cancel.
	if strings.Contains(msg.Payload, "jobCancel") {
		res := strings.Split(msg.Payload, ":")
		id := res[1]
		s.timeWheel.removeTask(id)
		return
	}
}
