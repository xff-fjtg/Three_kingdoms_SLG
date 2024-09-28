package data

import (
	"Three_kingdoms_SLG/server/game/model"
	"time"
)

// 数据库中的字段 不一定是客户端需要的字段
// 做业务逻辑的时候 会将数据库的结果 映射到客户端需要的结果上面
// 其中可能会做一些转换
type Role struct {
	RId        int       `gorm:"column:rid;primaryKey;autoIncrement"`
	UId        int       `gorm:"column:uid"`
	NickName   string    `gorm:"column:nick_name;size:20" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	Balance    int       `gorm:"column:balance"`
	HeadId     int16     `gorm:"column:headId"`
	Sex        int8      `gorm:"column:sex"`
	Profile    string    `gorm:"column:profile"`
	LoginTime  time.Time `gorm:"column:login_time"`
	LogoutTime time.Time `gorm:"column:logout_time"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (r *Role) TableName() string {
	return "role"
}

// 数据转换
func (r *Role) ToModel() interface{} {
	m := model.Role{}
	m.UId = r.UId
	m.RId = r.RId
	m.Sex = r.Sex
	m.NickName = r.NickName
	m.HeadId = r.HeadId
	m.Balance = r.Balance
	m.Profile = r.Profile
	return m
}
