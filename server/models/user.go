package models

import "time"

// User 结构体对应数据库中的用户表
type User struct {
	UId      int       `gorm:"column:uid;primaryKey;autoIncrement"`
	Username string    `gorm:"column:username;size:20;not null"`
	Passcode string    `gorm:"column:passcode;size:64"`
	Passwd   string    `gorm:"column:passwd;size:20;not null"`
	Hardware string    `gorm:"column:hardware;size:255"`
	Status   int       `gorm:"column:status"`
	Ctime    time.Time `gorm:"column:ctime;autoCreateTime"`
	Mtime    time.Time `gorm:"column:mtime;autoUpdateTime"`
	IsOnline bool      `gorm:"-"`
}

func (User) TableName() string {
	return "user"
}
