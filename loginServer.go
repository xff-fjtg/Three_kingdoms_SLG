package main

import (
	"Three_kingdoms_SLG/core"
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/login"
)

// http://localhost:8080/api/login
// websocket : ws://localhost:8080 服务器发消息 封装为路由
//
//localhost:8080 服务区 /api/login路由
func main() {
	core.InitConf()
	global.DB = core.InitGorm()
	host := global.Config.Login.Host
	port := global.Config.Login.Port
	s := net.InitServer(host + ":" + port)
	s.NeedSecret(false)
	login.Init()
	s.Router(login.Router)
	s.Start()
}
