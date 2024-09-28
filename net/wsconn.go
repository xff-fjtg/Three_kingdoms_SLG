package net

import "sync"

// request response
// 请求
type ReqBody struct {
	Seq   int64       `json:"seq"`
	Name  string      `json:"name"`
	Msg   interface{} `json:"msg"`
	Proxy string      `json:"proxy"`
}

// 回复
type RspBody struct {
	Seq  int64       `json:"seq"`
	Name string      `json:"name"`
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
}

type WsContext struct {
	mutex    sync.RWMutex
	property map[string]interface{}
}

type WsMsgReq struct {
	Body    *ReqBody
	Conn    WSConn
	Context *WsContext
}

type WsMsgRsp struct {
	Body *RspBody
}

func (ws *WsContext) Set(key string, value interface{}) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	ws.property[key] = value
}
func (ws *WsContext) Get(key string) interface{} {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()
	value, ok := ws.property[key]
	if ok {
		return value
	}
	return nil
}

// 理解为request请求 请求会有参数 请求中放参数 取参数
type WSConn interface {
	SetProperty(key string, value interface{})
	GetProperty(key string) (interface{}, error)
	RemoveProperty(key string)
	Addr() string
	Push(name string, data interface{})
}

const HandshakeMsg = "handshake"
const HeartbeatMsg = "heartbeat"

type Handshake struct {
	Key string `json:"key"`
}
type Heartbeat struct {
	CTime int64 `json:"ctime"`
	STime int64 `json:"stime"`
}
