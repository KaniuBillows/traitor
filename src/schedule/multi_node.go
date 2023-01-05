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
	"traitor/dao/model"
	"traitor/logger"
)

const (
	nodeId           = "traitor_defaultNode_"
	heartBeatTimeout = 1 //doesn't update node's status within 5s,indicating that this node is down. todo: with config.
	replicaCount     = 50
	redisChannel     = "traitor_pub_sub"
)

type MultiNodeSchedule struct {
	schedule
	client        *redis.Client
	NodeId        string
	cancel        context.CancelFunc
	consistentMap *consistenthash.Map
	cluster       string
}

func makeMultiNode(redisStr string, d dao.Dao, cluster string) *MultiNodeSchedule {
	opt, err := redis.ParseURL(redisStr)
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(opt)
	if err != nil {
		panic(err.Error())
	}
	uid := uuid.New()
	s := &MultiNodeSchedule{
		schedule: schedule{timeWheel: makeTimeWheel(), dao: d},
		client:   client,
		NodeId:   nodeId + cluster + uid.String(),
		cluster:  cluster,
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
	s.initJobs()
	s.startSub(ctx)
	s.timeWheel.start()
}

// load jobs from db.
func (s *MultiNodeSchedule) initJobs() {
	jbs, err := s.dao.GetRunnableJobs()
	if err != nil {
		panic(err)
	}
	for _, jb := range jbs {
		_ = s.addJob(&jb)
	}
}

func (s *MultiNodeSchedule) syncNodeList(ctx context.Context) {
	var cursor uint64 = 0
	list := make([]string, 0)
	matchStr := fmt.Sprintf("%s*", nodeId+s.cluster)
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
	consistenthash.RwLock.RLock()
	s.consistentMap = consistenthash.New(replicaCount, nil)
	s.consistentMap.Add(list...)
	consistenthash.RwLock.Unlock()
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

func (s *MultiNodeSchedule) handleAddJob(j *model.JobEntity) error {
	err := s.addJob(j)
	if err != nil {
		return err
	}
	err = s.notifyOtherNodes(fmt.Sprintf(jobAdd, j.JobId))
	if err != nil {
		return err
	}
	return nil
}

func (s *MultiNodeSchedule) notifyOtherNodes(content string) error {
	//notify other nodes.
	res := s.client.Publish(context.TODO(), redisChannel, content)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (s *MultiNodeSchedule) HandleJobStateChange(key string, state uint8) {
	if state == model.Runnable {
		job, err := s.dao.GetJobInfo(key)
		if err != nil {
			logger.Error(err)
			return
		}
		err = s.handleAddJob(&job)
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

func (s *MultiNodeSchedule) CreateTask(key string, execType uint8) func() {
	fn := s.schedule.CreateTask(key, execType)
	return func() {
		if s.executable(key) {
			fn()
		}
		return
	}
}

func (s *MultiNodeSchedule) executable(key string) bool {
	consistenthash.RwLock.RLock()
	defer func() {
		consistenthash.RwLock.RUnlock()
	}()
	return s.consistentMap.Get(key) == s.NodeId
}
func (s *MultiNodeSchedule) Remove(key string) {
	_ = s.cancelJob(key)
}

// cancelJob remove the job from the time wheel.
func (s *MultiNodeSchedule) cancelJob(key string) error {
	s.timeWheel.removeJob(key) // remove local job.
	// notify other nodes.
	err := s.notifyOtherNodes(fmt.Sprintf(jobCancel, key))
	if err != nil {
		return err
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
		err = s.addJob(&entity)
		if err != nil {
			logger.Error(fmt.Sprintf("handle sub jobAdd event error:%s", err.Error()))
		}
		return
	}

	// job cancel.
	if strings.Contains(msg.Payload, "jobCancel") {
		res := strings.Split(msg.Payload, ":")
		id := res[1]
		s.timeWheel.removeJob(id)
		return
	}
}

func (s *MultiNodeSchedule) HandleJobTimeChange(key string) {
	jb, err := s.dao.GetJobInfo(key)
	if err == nil {
		err = s.addJob(&jb)
		if err != nil {
			logger.Error(err)
		}
	} else {
		logger.Error(err)
	}
	// notify other nodes.
	err = s.notifyOtherNodes(fmt.Sprintf(fmt.Sprintf(jobAdd, key)))
	if err != nil {
		logger.Error(fmt.Sprintf("notify other nodes error: %s", err.Error()))
	}
}
