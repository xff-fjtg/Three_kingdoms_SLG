package controller

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/chat/chatModel"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/utils"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"log"
	"strings"
	"sync"
)

var GateHandler = &Handler{
	proxyMap: make(map[string]map[int64]*net.ProxyClient),
}

type Handler struct {
	proxyMutex sync.Mutex
	proxyMap   map[string]map[int64]*net.ProxyClient
	//map[string]map[int64]*net.ProxyClient
	//map[string]代理地址 后面的是 多个 客户端
	//代理地址-》客户端连接（游戏客户端id->连接）
	loginProxy string
	gameProxy  string
	chatProxy  string //聊天服务器
}

func (h *Handler) Router(r *net.Router) {
	h.loginProxy = global.Config.GateServer.LoginProxy
	h.gameProxy = global.Config.GateServer.GameProxy
	h.chatProxy = global.Config.GateServer.ChatProxy
	g := r.Group("*")
	g.AddRouter("*", h.all)
}

func (h *Handler) all(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	h.deal(req, rsp)
	//登陆聊天请求触发
	if req.Body.Name == "role.enterServer" && rsp.Body.Code == utils.OK {
		//进入游戏成功了 登录聊天服
		rspObj := &model.EnterServerRsp{}
		mapstructure.Decode(rsp.Body.Msg, rspObj)
		r := &chatModel.LoginReq{RId: rspObj.Role.RId, NickName: rspObj.Role.NickName, Token: rspObj.Token}
		reqBody := &net.ReqBody{Seq: 0, Name: "chat.login", Msg: r, Proxy: ""}
		rspBody := &net.RspBody{Seq: 0, Name: "chat.login", Msg: r, Code: 0}
		h.deal(&net.WsMsgReq{Body: reqBody, Conn: req.Conn}, &net.WsMsgRsp{Body: rspBody})
	}
}
func (h *Handler) deal(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	fmt.Println("gateServer handler")
	//代理转发
	name := req.Body.Name
	proxyStr := ""
	if isAccount(name) {
		proxyStr = h.loginProxy
	} else if isChat(name) {
		proxyStr = h.chatProxy
	} else {
		proxyStr = h.gameProxy
	}
	if proxyStr == "" { //代理服务器请求为空
		rsp.Body.Code = utils.ProxyNotInConnect
		return
	}
	h.proxyMutex.Lock()
	_, ok := h.proxyMap[proxyStr]
	if !ok {
		h.proxyMap[proxyStr] = make(map[int64]*net.ProxyClient)
	}
	h.proxyMutex.Unlock()
	c, err := req.Conn.GetProperty("cid")
	if err != nil {
		log.Println("cid is not exist", err)
		rsp.Body.Code = utils.InvalidParam
		return
	}
	cid := c.(int64)
	proxy := h.proxyMap[proxyStr][cid] //连接
	if proxy == nil {                  //第一次
		proxy = net.NewProxyClient(proxyStr)
		err = proxy.Connect()
		if err != nil {
			h.proxyMutex.Lock()
			delete(h.proxyMap[proxyStr], cid)
			h.proxyMutex.Unlock()
			rsp.Body.Code = utils.ProxyConnectError
			return
		}
		h.proxyMap[proxyStr][cid] = proxy
		proxy.SetProperty("cid", cid)
		proxy.SetProperty("proxy", proxyStr)
		proxy.SetProperty("gateConn", req.Conn)
		proxy.SetOnPush(h.onPush)
	}
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	r, err := proxy.Send(req.Body.Name, req.Body.Msg)
	if r != nil {
		rsp.Body.Code = r.Code
		rsp.Body.Msg = r.Msg
	} else {
		rsp.Body.Code = utils.ProxyConnectError
		return
	}
}

func (h *Handler) onPush(conn *net.ClientConn, body *net.RspBody) {
	gc, err := conn.GetProperty("gateConn")
	if err != nil {
		log.Println("onPush gateConn wrong", err)
		return
	}
	gateConn := gc.(net.WSConn)
	gateConn.Push(body.Name, body.Msg)

}

func isAccount(name string) bool {
	return strings.HasPrefix(name, "account.")
}
func isChat(name string) bool {
	return strings.HasPrefix(name, "chat.")
}
