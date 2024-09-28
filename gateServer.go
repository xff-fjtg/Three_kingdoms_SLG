package main

import (
	"Three_kingdoms_SLG/core"
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/gate"
)

/**
服务网关 = 路由转发+过滤器
路由转发：接收一切外界请求，转发到后端的服务上去。
过滤器：在服务网关中可以完成一系列的横切功能，例如权限校验、限流以及监控等。
引入网关之后，业务流程更改为：
1. 客户端请求地址：ws://127.0.0.1:8004
2. 当发起登录请求比如account.login时，转发请求到登录服务器（8003）处理
3. 当发起进入游戏请求 比如 role.enterServer时，转发请求的游戏服务器(8001)处理


1.登陆 account/login 要网关转发去登陆服务器
	1.game客户端 2.网关 3.login服务器
	1发请求2 2发给3 3相应2 2再相应1
2.网关作用（websocket的客户端）：如何和登陆服务器（websocket服务端）交互
3.网关又和game客户端交互，网关同时是websocket的服务端
4.websocket的服务端已经实现了
5.实现websocket的客户端
6.网关：代理服务器（保存代理地址 代理的连接通道） 客户端连接（websocket）链接
7.路由：接收所有的请求
8.要有握手协议 检测第一次建立连接的时候
*/

func main() {
	core.InitConf()
	global.DB = core.InitGorm()
	host := global.Config.GateServer.Host
	port := global.Config.GateServer.Port
	s := net.InitServer(host + ":" + port)
	s.NeedSecret(true)
	gate.Init()
	s.Router(gate.Router)
	s.Start()
}
