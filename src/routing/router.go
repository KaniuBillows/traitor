package routing

import "github.com/gin-gonic/gin"

func JobList(c *gin.Context) {

}

func RegistryRouting(engine *gin.Engine) {
	engine.GET("/jobList", JobList)
}
