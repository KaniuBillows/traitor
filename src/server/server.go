package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorhill/cronexpr"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"strings"
	"time"
	"traitor/dao"
	"traitor/dao/model"
	"traitor/schedule"
)

func (s *server) JobList(c *gin.Context) {

	jobs, err := s.dao.GetJobInfos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, jobs)
}
func (s *server) Remove(c *gin.Context) {
	id := c.Query("id")
	s.schedule.Remove(id)
	err := s.dao.RemoveJob(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

func (s *server) Update(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	var mp map[string]any
	var job model.JobEntity
	err := c.BindJSON(&mp)
	buffer, err := json.Marshal(mp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	err = json.Unmarshal(buffer, &job)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	delete(mp, model.State)
	delete(mp, model.LastExecTime)
	delete(mp, model.JobId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	entity, err := s.dao.GetJobInfo(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	var execType uint8
	// if job type would be changed.
	if _, ok := mp[model.ExecType]; ok {
		execType = job.ExecType
	} else {
		execType = entity.ExecType
	}

	// check time settings.
	err = checkTimeSettings(execType, job)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = s.dao.UpdateJob(id, mp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	s.schedule.HandleJobTimeChange(id)
	c.JSON(http.StatusOK, gin.H{})
}
func (s *server) Create(c *gin.Context) {
	execType, err := getJobType(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var job model.JobEntity
	err = c.BindJSON(&job)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	err = checkTimeSettings(execType, job)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job.State = model.Stop
	job.LastExecTime = nil
	job.ExecType = execType
	id, err := s.dao.AddJob(job)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": id,
	})
}
func (s *server) Start(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	entity, err := s.dao.GetJobInfo(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	enableStr := c.Query("enable")
	enable, err := strconv.ParseBool(enableStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	var runnable uint8
	if enable {
		err = checkTimeSettings(entity.ExecType, entity)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		runnable = model.Runnable
	} else {
		runnable = model.Stop
	}
	err = s.dao.UpdateJob(id, map[string]any{model.State: runnable})
	go s.schedule.HandleJobStateChange(id, runnable)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
func (s *server) UpdateScript(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	var mp map[string]string
	err := c.BindJSON(&mp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	sc, ok := mp["script"]
	if ok == false {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	err = s.dao.UpdateJob(id, map[string]any{model.Script: sc})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
}
func (s *server) EditPage(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusNotFound, gin.H{"err": "not found"})
		return
	}
	sc, _ := s.dao.GetJobScript(id)
	c.HTML(http.StatusOK, "edit_script.html", gin.H{
		"script": sc.Script,
	})
}

func (s *server) GetScript(c *gin.Context) {
	id := c.Param("id")
	sc, err := s.dao.GetJobScript(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"err": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sc})
}

func (s *server) GetJobInfo(c *gin.Context) {
	id := c.Query("id")
	j, err := s.dao.GetJobInfo(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": j})
}

func (s *server) Run(c *gin.Context) {
	execType, err := getJobType(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var entity model.JobEntity
	err = c.BindJSON(&entity)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	err = checkTimeSettings(execType, entity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.LastExecTime = nil
	entity.State = model.Runnable
	entity.ExecType = execType

	id, err := s.dao.AddJob(entity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	// schedule the job.
	s.schedule.HandleJobStateChange(id, model.Runnable)
	c.JSON(http.StatusOK, gin.H{
		"data": id,
	})
}
func checkTimeSettings(execType uint8, entity model.JobEntity) error {
	if execType == model.TimingExecute {
		// check cron
		err := cronCheck(entity.Cron)
		if err != nil {
			return errors.New("invalid cron expression")
		}
	} else {
		if entity.ExecAt == nil {
			return errors.New("invalid exec time")
		}
		// check delay time.
		if entity.ExecAt.ToTime().Sub(time.Now()) <= 5 {
			return errors.New("the minim delay is 5 seconds")
		}
	}
	return nil
}

func getJobType(c *gin.Context) (uint8, error) {
	t := c.Query("type")
	var execType uint8
	if "TIMING" == strings.ToUpper(t) {
		execType = model.TimingExecute
	} else if "DELAY" == strings.ToUpper(t) {
		execType = model.DelayExecute
	} else {
		return 0, errors.New("invalid job type")
	}
	return execType, nil
}

type wsWriter struct {
	ws *websocket.Conn
}

func (w *wsWriter) Write(p []byte) (n int, err error) {
	err = w.ws.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

/**********WS***********/

func (s *server) Debug(c *gin.Context) {
	id := c.Query("id")
	ws, err := s.upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	defer func(ws *websocket.Conn) {
		_ = ws.Close()
	}(ws)

	write := wsWriter{
		ws: ws,
	}

	fn, wt := s.schedule.CreateTaskForDebug(id, &write)
	go fn()
	wt.Wait()
}

/**********************/

type server struct {
	schedule schedule.Schedule
	dao      dao.Dao
	upgrade  websocket.Upgrader
}

var ser server

func makeServer() *server {
	s, d := schedule.StartStandalone()
	ser = server{
		schedule: s,
		dao:      d,
		upgrade: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	return &ser
}
func makeMultiServer(redisStr string, mongoUrl string, cluster string) *server {
	s, d := schedule.StartMultiNode(redisStr, mongoUrl, cluster)
	ser = server{
		schedule: s,
		dao:      d,
		upgrade: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	return &ser
}
func StartStandalone(engine *gin.Engine) {
	ser := makeServer()
	ser.RegistryRouting(engine)
	RegistryHtml(engine)
	ctx := context.Background()
	ser.schedule.Start(ctx)
}

func StartMultiNode(redisStr string, mongoUri string, cluster string, engine *gin.Engine) {
	ser := makeMultiServer(redisStr, mongoUri, cluster)
	ser.RegistryRouting(engine)
	RegistryHtml(engine)
	ctx := context.Background()
	ser.schedule.Start(ctx)
}

func cronCheck(cron string) error {
	_, err := cronexpr.Parse(cron)
	if err != nil {
		return errors.New("invalid cron expression")
	}
	return nil
}

func Close() {
	ser.schedule.Close()
}

func (s *server) RegistryRouting(engine *gin.Engine) {
	api := engine.Group("/api")
	{
		api.GET("/jobList", s.JobList)
		api.DELETE("/job", s.Remove)
		api.PUT("/job", s.Update)
		api.POST("/job", s.Create)
		api.POST("/script", s.UpdateScript)
		api.GET("/script", s.GetScript)
		api.GET("/debug", s.Debug)
		api.POST("/enable", s.Start)
		api.POST("/run", s.Run)
	}
	engine.GET("/edit/:id", s.EditPage)
}

func RegistryHtml(r *gin.Engine) {
	r.Static("/index", "./ui/html")
	r.Static("/js", "./ui/js")
	r.LoadHTMLGlob("ui/html/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

}
