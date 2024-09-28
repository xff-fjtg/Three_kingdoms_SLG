package main

import (
	"Three_kingdoms_SLG/core"
	"Three_kingdoms_SLG/global"
	mylog "Three_kingdoms_SLG/log"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/chat"
	"log"
)

func main() {
	core.InitConf()
	mylog.InitT1()
	host := global.Config.ChatServer.Host
	port := global.Config.ChatServer.Port
	s := net.InitServer(host + ":" + port)
	s.NeedSecret(false)
	chat.Init()
	s.Router(chat.Router)
	s.Start()
	log.Println("聊天服务启动成功")
}
