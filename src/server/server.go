package server

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorhill/cronexpr"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"traitor/dao"
	"traitor/dao/model"
	"traitor/logger"
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
	err := c.BindJSON(&mp)
	delete(mp, model.State)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	if _, ok := mp[model.Cron]; ok {
		_, tok := mp[model.Cron].(string)
		if tok == false {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cron expression"})
			return
		}
		err = cronCheck(mp[model.Cron].(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	err = s.dao.UpdateJob(id, mp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
func (s *server) Create(c *gin.Context) {
	var job model.JobEntity
	err := c.BindJSON(&job)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	job.State = model.Stop
	err = s.dao.AddJob(job)

	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{})
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
	err = cronCheck(entity.Cron)
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
	sc, err := s.dao.GetJobScript(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "edit_script.html", gin.H{
		"script": sc.Script,
	})
}

type wsWriter struct {
	ws *websocket.Conn
}

func (w *wsWriter) Write(p []byte) (n int, err error) {
	logger.Debug(p)
	err = w.ws.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

/**********WS***********/

func (s *server) Debug(c *gin.Context) {
	ws, err := s.upgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	defer func(ws *websocket.Conn) {
		_ = ws.Close()
	}(ws)

	_, message, err := ws.ReadMessage()
	if err != nil {
	}
	id := string(message)
	write := wsWriter{
		ws: ws,
	}

	fn, wt := s.schedule.CreateTaskForDebug(id, &write)
	fn()
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
func StartStandalone(engine *gin.Engine) {
	ser := makeServer()
	ser.RegistryRouting(engine)
	RegistryHtml(engine)
	ctx := context.Background()
	ser.schedule.Start(ctx)
}

func StartMultiNode(redisStr string, mongoUri string, cluster string, engine *gin.Engine) {
	ser := makeServer()
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
		api.GET("/debug", s.Debug)
		api.POST("/enable", s.Start)
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
