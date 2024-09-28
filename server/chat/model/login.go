package model

// 登陆聊天
type LoginReq struct {
	RId      int    `json:"rid"`
	NickName string `json:"nickName"`
	Token    string `json:"token"`
}

type LoginRsp struct {
	RId      int    `json:"rid"`
	NickName string `json:"nickName"`
}

// 加入聊天
type JoinReq struct {
	Type int8 `json:"type"` //0世界聊天、1联盟聊天
	Id   int  `json:"id"`
}
type JoinRsp struct {
	Type int8 `json:"type"` //0世界聊天、1联盟聊天
	Id   int  `json:"id"`
}

// 历史聊天
type HistoryReq struct {
	Type int8 `json:"type"` //0世界聊天、1联盟聊天
}
type HistoryRsp struct {
	Type int8      `json:"type"` //0世界聊天、1联盟聊天
	Msgs []ChatMsg `json:"msgs"`
}
type ChatMsg struct {
	RId      int    `json:"rid"`
	NickName string `json:"nickName"`
	Type     int8   `json:"type"` //0世界聊天、1联盟聊天
	Msg      string `json:"msg"`
	Time     int64  `json:"time"`
}

// 聊天
type ChatReq struct {
	Type int8   `json:"type"` //0世界聊天、1联盟聊天
	Msg  string `json:"msg"`
}

// 退出聊天
type ExitReq struct {
	Type int8 `json:"type"` //0世界聊天、1联盟聊天
	Id   int  `json:"id"`
}
type ExitRsp struct {
	Type int8 `json:"type"` //0世界聊天、1联盟聊天
	Id   int  `json:"id"`
}

// 注销
type LogoutReq struct {
	RId int `json:"RId"`
}
type LogoutRsp struct {
	RId int `json:"RId"`
}
