package logic

import (
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/redis"
	"Three_kingdoms_SLG/server/chat/model"
	"context"
	"encoding/json"
	"sync"
	"time"
)

// 聊天频道
type ChatGroup struct {
	userMutex sync.RWMutex
	msgMutex  sync.RWMutex
	//用户
	users map[int]*User
	//消息列表
	msgs ItemQueue
}

func (c *ChatGroup) Enter(user *User) {
	c.userMutex.Lock()
	defer c.userMutex.Unlock()
	c.users[user.rid] = user
}

func (c *ChatGroup) GetUser(rid int) *User {
	c.userMutex.Lock()
	defer c.userMutex.Unlock()
	return c.users[rid]
}
func (c *ChatGroup) Exit(rid int) {
	c.userMutex.Lock()
	defer c.userMutex.Unlock()
	delete(c.users, rid)
}

func (c *ChatGroup) History(t int8) []model.ChatMsg {
	//r := make([]model.ChatMsg, 0)
	msgs := c.msgs
	//从redis获取
	if t == 0 {
		//世界
		result, _ := redis.Pool.LRange(context.Background(), "chat_world", 0, -1).Result()
		for _, message := range result {
			msg := &Msg{}
			json.Unmarshal([]byte(message), msg)
			msgs.Enqueue(msg)
		}
	}
	//message列表
	c.msgs = msgs
	items := c.msgs.items
	chatMsgs := make([]model.ChatMsg, 0)
	for _, item := range items {
		msg := item.(*Msg)
		c := model.ChatMsg{RId: msg.RId, NickName: msg.NickName, Time: msg.Time.Unix(), Msg: msg.Msg}
		chatMsgs = append(chatMsgs, c)
	}

	return chatMsgs
}

func (c *ChatGroup) PutMsg(text string, rid int, t int8) *model.ChatMsg {
	c.userMutex.RLock()
	//拿到nickname
	u, ok := c.users[rid]
	c.userMutex.RUnlock()
	if ok == false {
		return nil
	}

	msg := &Msg{Msg: text, RId: rid, Time: time.Now(), NickName: u.nickName}
	//c.msgMutex.Lock()
	//size := c.msgs.Size()
	////消息太多 删掉一点
	//if size > 100 {
	//	c.msgs.Dequeue()
	//}
	////存消息 存到queue里面
	//c.msgs.Enqueue(msg)
	//c.msgMutex.Unlock()
	//redis list的数据结构 右进左出
	jsonMsg, _ := json.Marshal(msg)
	redis.Pool.RPush(context.Background(), "chat_world", jsonMsg)
	// 设置列表的生存时间为 1 小时（3600秒）
	redis.Pool.Expire(context.Background(), "chat_world", 7200*time.Second)
	//广播消息 给所有的用户广播
	c.userMutex.RLock()
	cm := &model.ChatMsg{RId: msg.RId, NickName: msg.NickName, Time: msg.Time.Unix(), Msg: msg.Msg, Type: t}
	for _, user := range c.users {
		net.Mgr.PushByRoleId(user.rid, "chat.push", cm)
	}
	c.userMutex.RUnlock()
	return cm
}
func NewGroup() *ChatGroup {
	return &ChatGroup{users: map[int]*User{}}
}
