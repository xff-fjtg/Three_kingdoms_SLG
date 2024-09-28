package main

import (
	"Three_kingdoms_SLG/core"
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/server/web"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func main() {
	core.InitConf()
	//global.DB = core.InitGorm()
	host := global.Config.WebServer.Host
	port := global.Config.WebServer.Port
	router := gin.Default()
	//路由
	web.Init(router)
	s := &http.Server{
		Addr:              host + ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	err := s.ListenAndServe()
	if err != nil {
		return
	}
}
