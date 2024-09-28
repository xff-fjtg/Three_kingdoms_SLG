package net

import (
	"Three_kingdoms_SLG/server/game/logic/conn"
	"Three_kingdoms_SLG/server/game/logic/pos"
	"github.com/gorilla/websocket"
	"sync"
)

var Mgr = NewMgr()

type WsMgr struct {
	uc sync.RWMutex
	cc sync.RWMutex
	rc sync.RWMutex

	userCache map[int]WSConn
	connCache map[int64]WSConn
	roleCache map[int]WSConn
}

func NewMgr() *WsMgr {
	return &WsMgr{
		userCache: make(map[int]WSConn),
		connCache: make(map[int64]WSConn),
		roleCache: make(map[int]WSConn),
	}
}

func (m *WsMgr) UserLogin(conn WSConn, uid int, token string) {
	m.uc.Lock()
	defer m.uc.Unlock()
	oldConn := m.userCache[uid]
	if oldConn != nil {
		//有用户登录着呢
		if conn != oldConn {
			//通过旧客户端 有用户抢登录了
			oldConn.Push("robLogin", nil)
		}
	}
	m.userCache[uid] = conn
	conn.SetProperty("uid", uid)
	conn.SetProperty("token", token)
}

// 为当前角色附上连接
func (w *WsMgr) RoleEnter(conn WSConn, rid int) {
	w.rc.Lock()
	defer w.rc.Unlock()
	conn.SetProperty("rid", rid)
	w.roleCache[rid] = conn
}
func (w *WsMgr) Push(pushSync conn.PushSync) {
	model := pushSync.ToModel()
	//找到所有对应的rid
	belongRIds := pushSync.BelongToRId()
	isCellView := pushSync.IsCellView()
	x, y := pushSync.Position()
	cells := make(map[int]int)
	//推送给开始位置
	if isCellView {
		//获得在当前位置所有的玩家 不管有多少玩家 只要看了当前位置 都要有信息返回
		cellRIds := pos.RPMgr.GetCellRoleIds(x, y, 8, 6)
		for _, rid := range cellRIds {
			//是否能出现在视野
			if can := pushSync.IsCanView(rid, x, y); can {
				w.PushByRoleId(rid, pushSync.PushMsgName(), model)
				cells[rid] = rid
			}
		}
	}
	//推送给目标位置 就是视野在这一块的玩家 这些位置对他们也要返回信息
	tx, ty := pushSync.TPosition()
	if tx >= 0 && ty >= 0 {
		var cellRIds []int
		if isCellView {
			//找视野内的玩家
			cellRIds = pos.RPMgr.GetCellRoleIds(tx, ty, 8, 6)
		} else {
			cellRIds = pos.RPMgr.GetCellRoleIds(tx, ty, 0, 0)
		}

		for _, rid := range cellRIds {
			if _, ok := cells[rid]; ok == false {
				if can := pushSync.IsCanView(rid, tx, ty); can {
					w.PushByRoleId(rid, pushSync.PushMsgName(), model)
					cells[rid] = rid
				}
			}
		}
	}

	//推送给自己
	for _, rid := range belongRIds {
		if _, ok := cells[rid]; ok == false {
			w.PushByRoleId(rid, pushSync.PushMsgName(), model)
		}
	}
}

func (w *WsMgr) PushByRoleId(rid int, msgName string, data interface{}) bool {
	if rid <= 0 {
		return false
	}
	w.rc.Lock()
	defer w.rc.Unlock()
	c, ok := w.roleCache[rid]
	if ok {
		c.Push(msgName, data)
		return true
	} else {
		return false
	}
}

var cid int64

// 这样连接就被管理起来了
func (w *WsMgr) NewConn(wsConn *websocket.Conn, needSecret bool) *WsServer {
	s := NewWsServer(wsConn, needSecret)
	cid++
	s.SetProperty("cid", cid)
	w.connCache[cid] = s
	return s
}

func (w *WsMgr) UserLogout(wsConn WSConn) {
	w.RemoveUser(wsConn)
}

func (w *WsMgr) RemoveUser(conn WSConn) {
	w.uc.Lock()
	//连接 缓存 还有一些东西 清除
	uid, err := conn.GetProperty("uid")
	if err == nil {
		//只删除自己的conn
		id := uid.(int)
		c, ok := w.userCache[id]
		if ok && c == conn {
			delete(w.userCache, id)
		}
	}
	w.uc.Unlock()

	w.rc.Lock()
	rid, err := conn.GetProperty("rid")
	if err == nil {
		//只删除自己的conn
		id := rid.(int)
		c, ok := w.roleCache[id]
		if ok && c == conn {
			delete(w.roleCache, id)
		}
	}
	w.rc.Unlock()

	conn.RemoveProperty("session")
	conn.RemoveProperty("uid")
	conn.RemoveProperty("role")
	conn.RemoveProperty("rid")
}
func NewWsServer(wsConn *websocket.Conn, needSecret bool) *WsServer {
	s := &WsServer{
		wsConn:     wsConn,
		outChan:    make(chan *WsMsgRsp, 1000),
		property:   make(map[string]interface{}),
		Seq:        0,
		needSecret: needSecret,
	}
	return s
}
