package model

type RegisterReq struct {
	Username string `form:"username" binding:"required" json:"username"`
	Password string `form:"password" binding:"required" json:"password"`
	Hardware string `form:"hardware" binding:"required" json:"hardware"`
}
