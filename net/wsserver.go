package net

import (
	"Three_kingdoms_SLG/utils"
	"encoding/json"
	"errors"
	"github.com/forgoer/openssl"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"log"
	"sync"
	"time"
)

// websocket 服务
type WsServer struct {
	wsConn       *websocket.Conn
	router       *Router
	outChan      chan *WsMsgRsp //需要写的信息写进去
	Seq          int64
	property     map[string]interface{}
	propertyLock sync.RWMutex //写属性进行加锁
	needSecret   bool
}

// 通道一旦建立，那么收发消息一定要监听
func (w *WsServer) Start() {
	//启动读写监听
	go w.readMsgLoop()
	go w.writeMsgLoop()
}
func (w *WsServer) readMsgLoop() {
	//先读到客户端发送过来的数据，然后 进行处理，然后在回消息
	//经过路由
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			w.Close()
		}
	}()
	for {
		_, data, err := w.wsConn.ReadMessage()
		if err != nil {
			log.Println("receiving message fail", err)
			break
		}
		//收到消息，解析json
		//1.data 解压
		data, err = utils.UnZip(data)
		if err != nil {
			log.Println("decompressing data fail", err)
			continue
		}
		//2.前端来的是加密信息 要解密
		if w.needSecret {
			secretKey, err := w.GetProperty("secretKey")
			if err == nil {
				//有加密
				key := secretKey.(string)
				//客户端传过来的数据是加密的 需要解密
				d, err := utils.AesCBCDecrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
				if err != nil {
					log.Println("数据格式有误，解密失败:", err)
					//出错后 发起握手
					//两边密钥看不一样
					w.Handshake()
				} else {
					data = d
				}
			}
		}
		//3.data->body
		body := &ReqBody{}

		err = json.Unmarshal(data, body) //拿到请求数据了
		if err != nil {
			log.Println("change message fail，数据格式非法", err)
			break
		} else {
			context := &WsContext{
				property: make(map[string]interface{}),
			}
			//获取到前端的数据了，那么要拿上数据去处理了
			req := &WsMsgReq{Conn: w, Body: body, Context: context}
			rsp := &WsMsgRsp{Body: &RspBody{Name: body.Name, Seq: req.Body.Seq}}
			if req.Body.Name == HeartbeatMsg {
				//收到心跳请求要回消息
				h := &Heartbeat{}
				err = mapstructure.Decode(req.Body.Msg, h)
				if err != nil {
					log.Println("decode heartbeat fail", err)
					return
				}
				h.STime = time.Now().UnixNano() / 1e6
				rsp.Body.Msg = h

			} else {
				if w.router != nil {
					w.router.Run(req, rsp)
				}
			}
			w.outChan <- rsp
		}
	}
	w.Close()
}

// send message
func (w *WsServer) writeMsgLoop() {
	for {
		select {
		case msg := <-w.outChan:
			w.Write(msg)
		}
	}
}
func (w *WsServer) Write(msg *WsMsgRsp) {
	data, err := json.Marshal(msg.Body) // 将消息转换为 JSON 格式
	if err != nil {
		log.Println(err)
	}
	//2.前端来的是加密信息 要解密
	secretKey, err := w.GetProperty("secretKey")
	if err == nil {
		//有加密
		key := secretKey.(string)
		//数据加密
		data, _ = utils.AesCBCEncrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
	}
	if data, err = utils.Zip(data); err == nil {
		err := w.wsConn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			log.Println("Server write message fail", err)
			return
		}
		d, _ := json.Marshal(msg)
		//mylog.DefaultLog.Info("服务端收到消息并发送: ", zap.String("body", string(d)))
		log.Println("服务端写数据", string(d))
	}
}

// Close
func (w *WsServer) Close() {
	_ = w.wsConn.Close()
}

// Handshake 握手协议
// 当游戏客户端发送请求 会先要握手
// 后端会发送对应的加密key给客户端
// 客户端在发送数据的时候就会用这个key进行加密处理
// 断开连接 一定要握手
func (w *WsServer) Handshake() {
	secretKey := ""
	key, err := w.GetProperty("secretKey")
	if err == nil {
		secretKey = key.(string)
	} else {
		secretKey = utils.RandSeq(16)
	}
	handshake := &Handshake{Key: secretKey}
	body := &RspBody{Name: HandshakeMsg, Msg: handshake}
	if data, err := json.Marshal(body); err == nil {
		if secretKey != "" {
			w.SetProperty("secretKey", secretKey)
		} else {
			w.RemoveProperty("secretKey")
		}
		if data, err = utils.Zip(data); err == nil {
			err := w.wsConn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				log.Println("Server write message fail", err)
				return
			}
		}
	}
}

func (w *WsServer) Router(router *Router) {
	w.router = router
}
func (w *WsServer) GetProperty(key string) (interface{}, error) {
	w.propertyLock.RLock()
	defer w.propertyLock.RUnlock()
	if value, ok := w.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}
func (w *WsServer) RemoveProperty(key string) {
	w.propertyLock.Lock()
	defer w.propertyLock.Unlock()
	delete(w.property, key)
}
func (w *WsServer) Addr() string {
	return w.wsConn.RemoteAddr().String()
}
func (w *WsServer) Push(name string, data interface{}) {
	rsp := &WsMsgRsp{Body: &RspBody{Name: name, Msg: data, Seq: 0}}
	w.outChan <- rsp
}
func (w *WsServer) SetProperty(key string, value interface{}) {
	w.propertyLock.Lock()
	defer w.propertyLock.Unlock()
	w.property[key] = value
}
