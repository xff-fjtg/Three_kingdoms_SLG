package net

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Server struct {
	addr       string
	router     *Router
	needSecret bool //网关和客户端加密  其他不加密
}

func InitServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}
func (s *Server) NeedSecret(needSecret bool) {
	s.needSecret = needSecret
}
func (s *Server) Router(router *Router) {
	s.router = router
}

// 启动服务
func (s *Server) Start() {
	http.HandleFunc("/", s.wsHandler)
	err := http.ListenAndServe(s.addr, nil)
	if err != nil {
		panic(err)
	}
}

// http upgrade websocket
var wsUpgrader = websocket.Upgrader{
	// allow cors request
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 路由
//
//	w http.ResponseWriter：这是 HTTP 响应的接口，你可以通过它向客户端发送响应。
//	r *http.Request：这是 HTTP 请求的结构体，包含了请求的所有信息
func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	//web socket
	//1.http upgrade->websocket
	wsConn, err := wsUpgrader.Upgrade(w, r, nil)
	//这是一个 *websocket.Conn 对象，表示成功建立的 WebSocket 连接。你可以通过这个连接对象来收发 WebSocket 消息。
	if err != nil {
		log.Fatal("websocket server fail", err)
	}
	log.Println("websocket connect success")
	//无论客户端还是服务端都能收发消息
	//发消息要把消息当路由处理，要先定义消息格式

	wsServer := Mgr.NewConn(wsConn, s.needSecret)
	wsServer.Router(s.router)
	wsServer.Start()
	wsServer.Handshake()
}
