package controller

import (
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/chat/logic"
	"Three_kingdoms_SLG/server/chat/middleware"
	"Three_kingdoms_SLG/server/chat/model"
	"Three_kingdoms_SLG/utils"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"log"
	"sync"
)

var ChatController = &chatController{
	worldGroup:       logic.NewGroup(),
	unionGroups:      make(map[int]*logic.ChatGroup),
	ridToUnionGroups: make(map[int]int),
}

type chatController struct {
	unionMutex sync.RWMutex

	worldGroup       *logic.ChatGroup         //世界频道
	unionGroups      map[int]*logic.ChatGroup //联盟频道
	ridToUnionGroups map[int]int              //rid对应的联盟频道id
}

func (c *chatController) Router(router *net.Router) {
	g := router.Group("chat")
	g.Use(middleware.Log())
	//看看为什么不能checkrole
	g.AddRouter("login", c.login)
	g.AddRouter("join", c.join)
	g.AddRouter("history", c.history) //历史记录
	g.AddRouter("chat", c.chat)
	g.AddRouter("exit", c.exit)     //退出
	g.AddRouter("logout", c.logout) //聊天注销
}

func (c *chatController) login(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//登陆聊天室
	//登陆进去 所有玩家都可以在世界频道聊天
	reqObj := &model.LoginReq{}
	rspObj := &model.LoginRsp{}
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj

	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj.RId = reqObj.RId
	rspObj.NickName = reqObj.NickName
	//登陆是否合法
	_, _, err := utils.ParseToken(reqObj.Token)
	if err != nil {
		rsp.Body.Code = utils.InvalidParam
		return
	}
	//登陆
	net.Mgr.RoleEnter(req.Conn, reqObj.RId)
	//加入用户列表
	c.worldGroup.Enter(logic.NewUser(reqObj.RId, reqObj.NickName))
	fmt.Println("yes!!!")
}

func (c *chatController) join(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.JoinReq{}
	rspObj := &model.JoinRsp{}
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj
	rspObj.Type = reqObj.Type
	mapstructure.Decode(req.Body.Msg, reqObj)
	p, _ := req.Conn.GetProperty("rid")
	rid := p.(int)
	//联盟聊天
	if reqObj.Type == 1 {

		u := c.worldGroup.GetUser(rid)
		if u == nil {
			rsp.Body.Code = utils.InvalidParam
			return
		}

		c.unionMutex.Lock()
		//拿到联盟聊天的id
		gId, ok := c.ridToUnionGroups[rid]
		if ok {
			if gId != reqObj.Id {
				//现存的id和旧的不一样 联盟聊天只能有一个，顶掉旧的
				if g, ok := c.unionGroups[gId]; ok {
					//删除
					g.Exit(rid)
				}
				_, ok = c.unionGroups[reqObj.Id]
				if ok == false {
					c.unionGroups[reqObj.Id] = logic.NewGroup()
				}
				c.ridToUnionGroups[rid] = reqObj.Id
				c.unionGroups[reqObj.Id].Enter(u)
			}
		} else {
			//未加入联盟频道 要新加入
			_, ok = c.unionGroups[reqObj.Id]
			if ok == false {
				//创建新组
				c.unionGroups[reqObj.Id] = logic.NewGroup()
			}
			c.ridToUnionGroups[rid] = reqObj.Id
			c.unionGroups[reqObj.Id].Enter(u)
		}
		c.unionMutex.Unlock()
	}
	fmt.Println("yes!!!")
}

func (c *chatController) history(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.HistoryReq{}
	rspObj := &model.HistoryRsp{}
	rsp.Body.Code = utils.OK

	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj.Msgs = []model.ChatMsg{}
	p, _ := req.Conn.GetProperty("rid")
	rid := p.(int)

	if reqObj.Type == 0 {
		//世界聊天
		r := c.worldGroup.History(0)
		rspObj.Msgs = r
	} else if reqObj.Type == 1 {
		//联盟聊天
		c.unionMutex.RLock()
		//先找到联盟id
		id, ok := c.ridToUnionGroups[rid]
		if ok {
			g, ok := c.unionGroups[id]
			if ok {
				rspObj.Msgs = g.History(1)
			}
		}
		c.unionMutex.RUnlock()
	}
	rspObj.Type = reqObj.Type
	rsp.Body.Msg = rspObj
}

func (c *chatController) chat(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.ChatReq{}
	rspObj := &model.ChatMsg{}
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj

	mapstructure.Decode(req.Body.Msg, reqObj)

	p, _ := req.Conn.GetProperty("rid")
	rid := p.(int)
	if reqObj.Type == 0 {
		//世界聊天
		rsp.Body.Msg = c.worldGroup.PutMsg(reqObj.Msg, rid, 0)
	} else if reqObj.Type == 1 {
		//前端有个判断 无联盟 无法联盟聊天
		//联盟聊天
		c.unionMutex.RLock()
		//拿到联盟id
		id, ok := c.ridToUnionGroups[rid]
		if ok {
			g, ok := c.unionGroups[id]
			if ok {
				g.PutMsg(reqObj.Msg, rid, 1)
			} else {
				log.Println("chat not found rid in unionGroups", rid)
			}
		} else {
			log.Println("chat not found rid in ridToUnionGroups", rid)
		}
		c.unionMutex.RUnlock()
	}
}

func (c *chatController) exit(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.ExitReq{}
	rspObj := &model.ExitRsp{}
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj
	rspObj.Type = reqObj.Type
	mapstructure.Decode(req.Body.Msg, reqObj)
	p, _ := req.Conn.GetProperty("rid")
	rid := p.(int)

	if reqObj.Type == 1 {
		//退出联盟要退出联盟聊天
		c.unionMutex.Lock()
		id, ok := c.ridToUnionGroups[rid]
		if ok {
			g, ok := c.unionGroups[id]
			if ok {
				g.Exit(rid)
			}
		}
		//退出
		delete(c.ridToUnionGroups, rid)
		c.unionMutex.Unlock()
	}
}

func (c *chatController) logout(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.LogoutReq{}
	rspObj := &model.LogoutRsp{}
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj

	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj.RId = reqObj.RId

	net.Mgr.UserLogout(req.Conn)
	//退出
	c.worldGroup.Exit(reqObj.RId)
}
