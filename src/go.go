package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"traitor/server"
)

func main() {

	r := gin.Default()

	var mode string // running mode
	var redisUri string
	var mongoStr string
	var cluster string //cluster name.
	var ip string
	var port int
	flag.StringVar(&mode, "m", "std", "[std] or [multi] running mode,default is std for standalone server.")
	flag.StringVar(&redisUri, "r", "", "redis connection string.required for multi mode.")
	flag.StringVar(&mongoStr, "mg", "", "mongodb uri.required for multi mode.")
	flag.StringVar(&cluster, "c", "", "multi nodes cluster name.only effective for multi mode.")
	flag.StringVar(&ip, "ip", "", "bind ip address.default is empty for all address.")
	flag.IntVar(&port, "p", 8080, "bind port")
	flag.Parse()
	if mode == "multi" {
		if redisUri == "" {
			panic("redis address is required.")
		}
		server.StartMultiNode(redisUri, mongoStr, cluster, r)
	} else {
		server.StartStandalone(r)

	}

	r.NoRoute(func(ctx *gin.Context) { ctx.JSON(http.StatusNotFound, gin.H{}) })
	defer server.Close()

	err := r.Run(fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Println(err)
		return
	}
}
