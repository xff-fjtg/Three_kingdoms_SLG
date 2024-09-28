package data

import (
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/game/model"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
)

type Skill struct {
	Id             int    `gorm:"column:id;primaryKey;autoIncrement"`
	RId            int    `gorm:"column:rid"`
	CfgId          int    `gorm:"column:cfgId"`
	BelongGenerals string `gorm:"column:belong_generals"`
	Generals       []int  `gorm:"-"`
}

func NewSkill(rid int, cfgId int) *Skill {
	return &Skill{
		CfgId:          cfgId,
		RId:            rid,
		Generals:       []int{},
		BelongGenerals: "[]",
	}
}

func (s *Skill) TableName() string {
	return "skill"
}

func (s *Skill) ToModel() interface{} {
	p := model.Skill{}
	p.Id = s.Id
	p.CfgId = s.CfgId
	p.Generals = s.Generals
	return p
}

func (a *Skill) AfterFind(tx *gorm.DB) (err error) {
	// 假设 belong_generals 是数据库中的字段名
	if tx.Statement.Schema.LookUpField("BelongGenerals") != nil {
		a.Generals = make([]int, 0)
		if len(a.BelongGenerals) > 0 { // 假设 BelongGenerals 是存储数据的字段
			err := json.Unmarshal([]byte(a.BelongGenerals), &a.Generals)
			if err != nil {
				return err
			}
			fmt.Println(a.Generals)
		}
	}
	return nil
}

/* 推送同步 begin */
func (s *Skill) IsCellView() bool {
	return false
}

func (s *Skill) IsCanView(rid, x, y int) bool {
	return false
}

func (s *Skill) BelongToRId() []int {
	return []int{s.RId}
}

func (s *Skill) PushMsgName() string {
	return "skill.push"
}

func (s *Skill) Position() (int, int) {
	return -1, -1
}

func (s *Skill) TPosition() (int, int) {
	return -1, -1
}

func (s *Skill) Push() {
	net.Mgr.Push(s)
}
