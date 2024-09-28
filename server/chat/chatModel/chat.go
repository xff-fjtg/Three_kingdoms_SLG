package chatModel

type LoginReq struct {
	RId      int    `json:"rid"`
	NickName string `json:"nickName"`
	Token    string `json:"token"`
}

type LoginRsp struct {
	RId      int    `json:"rid"`
	NickName string `json:"nickName"`
}
