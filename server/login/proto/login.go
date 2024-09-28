package proto

// 用户发送的消息
type LoginRsp struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Session  string `json:"session"`
	UId      int    `json:"uid"`
}

// 前端传递过来的消息
type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Ip       string `json:"ip"`
	Hardware string `json:"hardware"`
}

// 重新登陆
type ReLoginReq struct {
	Session  string `json:"session"`
	Ip       string `json:"ip"`
	Hardware string `json:"hardware"`
}
type ReLoginRsp struct {
	Session string `json:"session"`
}
