package main

import (
	"github.com/joeycumines/kubestatus.go"
	"time"
	"github.com/gin-gonic/gin"
	"fmt"
)

var config kubestatus.Config

func main() {
	server, err := kubestatus.NewService(config)
	config.GinHandlers = []gin.HandlerFunc{
		gin.Logger(),
		gin.Recovery(),
	}
	config.StartWait = time.Millisecond * 10
	if err != nil {
		panic(fmt.Errorf("%+v", err))
	}
	server.Start()
	<-server.Ctx().Done()
}
