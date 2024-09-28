package model

import "time"

// LoginHistory 结构体对应数据库中的登录历史表
type LoginHistory struct {
	Id       int       `gorm:"column:id;primaryKey;autoIncrement"`
	UId      int       `gorm:"column:uid"`
	CTime    time.Time `gorm:"column:ctime"`
	Ip       string    `gorm:"column:ip;size:45"`
	State    int8      `gorm:"column:state;default:0"`
	Hardware string    `gorm:"column:hardware;size:255"`
}

// LoginLast 结构体对应数据库中的最后登录信息表
type LoginLast struct {
	Id         int        `gorm:"column:id;primaryKey;autoIncrement"`
	UId        int        `gorm:"column:uid"`
	LoginTime  *time.Time `gorm:"column:login_time"`
	LogoutTime *time.Time `gorm:"column:logout_time"`
	Ip         string     `gorm:"column:ip;size:45"` // 假设 IPv6 地址, 大小设为 45
	Session    string     `gorm:"column:session;size:255"`
	IsLogout   int8       `gorm:"column:is_logout;default:0"`
	Hardware   string     `gorm:"column:hardware;size:255"`
}

func (LoginHistory) TableName() string {
	return "login_history"
}

func (LoginLast) TableName() string {
	return "login_last"
}

const (
	Login = iota
	Logout
)
