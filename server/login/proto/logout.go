package proto

// 登出
type LogoutReq struct {
	UId int `json:"uid"`
}
type LogoutRsp struct {
	UId int `json:"uid"`
}
