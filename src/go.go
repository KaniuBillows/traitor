package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"traitor/routing"
)

func main() {
	r := gin.Default()
	routing.RegistryRouting(r)
	r.NoRoute(func(ctx *gin.Context) { ctx.JSON(http.StatusNotFound, gin.H{}) })
	err := r.Run(":8080")
	if err != nil {
		fmt.Println(err)
		return
	}
}
