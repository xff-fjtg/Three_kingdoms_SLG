package main

import (
	"Three_kingdoms_SLG/core"
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/game"
)

/**
1.登陆完成，创建角色
2.根据用户有的角色进去游戏，没有就创建
3.有乱七八糟的东西 钱 资源 武将 这些数据 要查询 展示出来
4.地图什么的 全局资源 要定义
5.资源 军队 城池  武将 加载出来
6.
*/

func main() {

	core.InitConf()
	//global.DB = core.InitGorm()
	host := global.Config.GameServer.Host
	port := global.Config.GameServer.Port
	s := net.InitServer(host + ":" + port)
	s.NeedSecret(false)
	game.Init()
	s.Router(game.Router)
	s.Start()
}
