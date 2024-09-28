package net

import (
	"Three_kingdoms_SLG/utils"
	"context"
	"encoding/json"
	"errors"
	"github.com/forgoer/openssl"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"log"
	"sync"
	"time"
)

type ClientConn struct {
	WsConn        *websocket.Conn
	isClosed      bool //监听当前客户端是否关闭
	property      map[string]interface{}
	propertyLock  sync.RWMutex
	Seq           int64
	handshake     bool
	handshakeChan chan bool
	onPush        func(conn *ClientConn, body *RspBody)
	onClose       func(conn *ClientConn)
	syncCtxMap    map[int64]*syncCtx
	syncCtxLock   sync.RWMutex
}
type syncCtx struct {
	//Goroutine 的上下文，包含 Goroutine 的运行状态、环境、现场等信息
	ctx     context.Context
	cancel  context.CancelFunc
	outChan chan *RspBody
}

func (w *ClientConn) Start() bool {
	//一直不停接收消息
	//等待握手的消息返回
	w.handshake = false
	go w.wsReadLoop()
	return w.waitHandShake()
}
func (s *syncCtx) wait() *RspBody {
	//受消息
	select {
	case msg := <-s.outChan:
		return msg
	case <-s.ctx.Done():
		log.Println("proxy server timeout")
		return nil
	}
}
func (w *ClientConn) waitHandShake() bool {
	//等待握手消息
	//万一程序超时 响应不到
	cxt, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	select {
	case _ = <-w.handshakeChan:
		log.Println("握手成功")
		return true
	case <-cxt.Done():
		log.Println("握手超时")
		return false
	}
}

func (w *ClientConn) wsReadLoop() {
	//for {
	//	_, data, err := c.WsConn.ReadMessage()
	//	fmt.Println(data, err)
	//	//可能读到 握手 心跳 请求信息(要对应处理)
	//	//收到握手消息
	//	c.handshake = true
	//	c.handshakeChan <- true
	//}
	//先读到客户端 发送过来的数据，然后 进行处理，然后在回消息
	//经过路由 实际处理程序
	defer func() {
		if err := recover(); err != nil {
			log.Println("服务端捕捉到异常", err)
			w.Close()
		}
	}()
	for {
		_, data, err := w.WsConn.ReadMessage()
		if err != nil {
			log.Println("收消息出现错误:", err)
			break
		}
		//收到消息 解析消息 前端发送过来的消息 就是json格式
		//1. data 解压 unzip
		data, err = utils.UnZip(data)
		if err != nil {
			log.Println("解压数据出错，非法格式：", err)
			continue
		}
		//2. 前端的消息 加密消息 进行解密

		secretKey, err := w.GetProperty("secretKey")
		if err == nil {
			//有加密
			key := secretKey.(string)
			//客户端传过来的数据是加密的 需要解密
			d, err := utils.AesCBCDecrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
			if err != nil {
				log.Println("数据格式有误，解密失败:", err)
			} else {
				data = d
			}
		}
		//3. data 转为body
		body := &RspBody{}
		err = json.Unmarshal(data, body)
		if err != nil {
			log.Println("服务端json格式解析有误，非法格式:", err)
		} else {
			//判断是握手还是别的什么
			if body.Seq == 0 {
				if body.Name == HandshakeMsg {
					//获取密钥
					hs := &Handshake{}
					mapstructure.Decode(body.Msg, hs)
					if hs.Key != "" {
						w.SetProperty("secretKey", hs.Key)
					} else {
						w.RemoveProperty("secretKey")
					}
					log.Println("handshake")
					w.handshake = true
					w.handshakeChan <- true
				} else {
					if w.onPush != nil {
						w.onPush(w, body)
					}
				}
			} else {
				w.syncCtxLock.RLock()
				ctx, ok := w.syncCtxMap[body.Seq]
				w.syncCtxLock.RUnlock()
				if ok {
					ctx.outChan <- body
				} else {
					log.Println("no seq syncCtx find")
				}
			}
		}
	}
	w.Close()
}

func (w *ClientConn) Close() {
	_ = w.WsConn.Close()
}

func NewClientConn(wsConn *websocket.Conn) *ClientConn {
	return &ClientConn{
		WsConn:        wsConn,
		handshakeChan: make(chan bool),
		Seq:           0,
		isClosed:      false,
		property:      make(map[string]interface{}),
		syncCtxMap:    map[int64]*syncCtx{},
	}
}
func NewSynCtx() *syncCtx {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	return &syncCtx{
		ctx:     ctx,
		cancel:  cancel,
		outChan: make(chan *RspBody),
	}
}

func (w *ClientConn) RemoveProperty(key string) {
	w.propertyLock.Lock()
	defer w.propertyLock.Unlock()
	delete(w.property, key)
}
func (w *ClientConn) Addr() string {
	return w.WsConn.RemoteAddr().String()
}
func (w *ClientConn) Push(name string, data interface{}) {
	rsp := &WsMsgRsp{Body: &RspBody{Name: name, Msg: data, Seq: 0}}
	//w.OutChan <- rsp
	w.write(rsp.Body)
}
func (w *ClientConn) SetProperty(key string, value interface{}) {
	w.propertyLock.Lock()
	defer w.propertyLock.Unlock()
	w.property[key] = value
}
func (w *ClientConn) GetProperty(key string) (interface{}, error) {
	w.propertyLock.RLock()
	defer w.propertyLock.RUnlock()
	if value, ok := w.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

func (w *ClientConn) write(body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		log.Println(err)
	}
	//2.前端来的是加密信息 要解密
	//secretKey, err := w.GetProperty("secretKey")
	//if err == nil {
	//	//有加密
	//	key := secretKey.(string)
	//	//数据加密
	//	data, err = utils.AesCBCEncrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
	//	if err != nil {
	//		log.Println("Encryption failed")
	//		return err
	//	}
	//}
	if data, err = utils.Zip(data); err == nil {
		err := w.WsConn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			log.Println("Server write message fail", err)
			return err
		}
	} else {
		log.Println("Failed to compress data")
		return err
	}
	return nil
}

func (w *ClientConn) SetOnPush(hook func(conn *ClientConn, body *RspBody)) {
	w.onPush = hook
}

func (w *ClientConn) Send(name string, msg interface{}) (*RspBody, error) {
	//把请求 发送给 代理服务器（登陆服务器） 然后等待返回
	w.Seq += 1
	seq := w.Seq
	sc := NewSynCtx()
	w.syncCtxLock.Lock()
	req := &ReqBody{Name: name, Msg: msg, Seq: seq}
	w.syncCtxMap[seq] = sc
	w.syncCtxLock.Unlock()
	rsp := &RspBody{Name: name, Seq: seq, Code: utils.OK}
	err := w.write(req)
	if err != nil {
		sc.cancel()
	} else {
		r := sc.wait()
		if r == nil {
			rsp.Code = utils.ProxyConnectError
		} else {
			rsp = r
		}
	}
	w.syncCtxLock.Lock()
	delete(w.syncCtxMap, seq)
	w.syncCtxLock.Unlock()
	return rsp, nil

}
