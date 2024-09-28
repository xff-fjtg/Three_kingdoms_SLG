package data

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/game/gameConfig/general"
	"Three_kingdoms_SLG/server/game/model"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"time"
)

const (
	GeneralNormal      = 0 //正常
	GeneralComposeStar = 1 //星级合成
	GeneralConvert     = 2 //转换
)
const SkillLimit = 3

type General struct {
	Id            int             `gorm:"column:id;primaryKey;autoIncrement"`
	RId           int             `gorm:"column:rid"`
	CfgId         int             `gorm:"column:cfgId"`
	PhysicalPower int             `gorm:"column:physical_power"`
	Level         int8            `gorm:"column:level"`
	Exp           int             `gorm:"column:exp"`
	Order         int8            `gorm:"column:order"`
	CityId        int             `gorm:"column:cityId"`
	CreatedAt     time.Time       `gorm:"column:created_at"`
	CurArms       int             `gorm:"column:arms"`
	HasPrPoint    int             `gorm:"column:has_pr_point"`
	UsePrPoint    int             `gorm:"column:use_pr_point"`
	AttackDis     int             `gorm:"column:attack_distance"`
	ForceAdded    int             `gorm:"column:force_added"`
	StrategyAdded int             `gorm:"column:strategy_added"`
	DefenseAdded  int             `gorm:"column:defense_added"`
	SpeedAdded    int             `gorm:"column:speed_added"`
	DestroyAdded  int             `gorm:"column:destroy_added"`
	StarLv        int8            `gorm:"column:star_lv"`
	Star          int8            `gorm:"column:star"`
	ParentId      int             `gorm:"column:parentId"`
	Skills        string          `gorm:"column:skills"`
	SkillsArray   []*model.GSkill `gorm:"-"` // 这个字段不映射到数据库
	State         int8            `gorm:"column:state"`
}

// 如果需要自定义表名

func (g *General) TableName() string {
	return "general"
}

func (g *General) ToModel() interface{} {
	p := model.General{}
	p.CityId = g.CityId
	p.Order = g.Order
	p.PhysicalPower = g.PhysicalPower
	p.Id = g.Id
	p.CfgId = g.CfgId
	p.Level = g.Level
	p.Exp = g.Exp

	p.CurArms = g.CurArms
	p.HasPrPoint = g.HasPrPoint
	p.UsePrPoint = g.UsePrPoint
	p.AttackDis = g.AttackDis
	p.ForceAdded = g.ForceAdded
	p.StrategyAdded = g.StrategyAdded
	p.DefenseAdded = g.DefenseAdded
	p.SpeedAdded = g.SpeedAdded
	p.DestroyAdded = g.DestroyAdded
	p.StarLv = g.StarLv
	p.Star = g.Star
	p.State = g.State
	p.ParentId = g.ParentId
	p.Skills = g.SkillsArray
	return p
}

var GenerDao = &generDao{
	genChan: make(chan *General, 100),
}

type generDao struct {
	genChan chan *General
}

func (g *generDao) running() {
	for {
		select {
		case gen := <-g.genChan:
			if gen.Id > 0 && gen.RId > 0 {
				global.DB.Model(gen).Select(
					"level", "exp", "order", "cityId",
					"physical_power", "star_lv", "has_pr_point",
					"use_pr_point", "force_added", "strategy_added",
					"defense_added", "speed_added", "destroy_added",
					"parentId", "compose_type", "skills", "state",
				).Where("id =?", gen.Id).Updates(&gen)

			}
		}
	}
}
func (g *General) SyncExecute() {
	GenerDao.genChan <- g
	//推送
	g.Push()
}

func init() {
	go GenerDao.running()
}
func (g *General) GetDestroy() int {
	cfg, ok := general.General.GMap[g.CfgId] //获取general detail
	if ok {
		return cfg.Destroy + cfg.DestroyGrow*int(g.Level) + g.DestroyAdded
	}
	return 0
}
func (g *General) GetForce() int {
	cfg, ok := general.General.GMap[g.CfgId] //获取general detail
	if ok {
		return cfg.Force + cfg.ForceGrow*int(g.Level) + g.ForceAdded
	}
	return 0
}
func (g *General) GetSpeed() int {
	cfg, ok := general.General.GMap[g.CfgId] //获取general detail
	if ok {
		return cfg.Speed + cfg.SpeedGrow*int(g.Level) + g.SpeedAdded
	}
	return 0
}
func (g *General) GetStrategy() int {
	cfg, ok := general.General.GMap[g.CfgId] //获取general detail
	if ok {
		return cfg.Strategy + cfg.StrategyGrow*int(g.Level) + g.StrategyAdded
	}
	return 0
}

func (g *General) GetDefense() int {
	cfg, ok := general.General.GMap[g.CfgId] //获取general detail
	if ok {
		return cfg.Defense + cfg.DefenseGrow*int(g.Level) + g.DefenseAdded
	}
	return 0
}

func (g *General) AfterSet(tx *gorm.DB) {
	// 检查字段 "skills"
	if g.Skills != "" {
		g.SkillsArray = make([]*model.GSkill, 3)
		// 尝试反序列化 "skills" 字段为 g.SkillsArray
		err := json.Unmarshal([]byte(g.Skills), &g.SkillsArray)
		if err != nil {
			fmt.Println("Error unmarshaling skills:", err)
		}
		fmt.Println(g.SkillsArray)
	}
}

// beforeModify 用于将 SkillsArray 序列化为 Skills
func (g *General) beforeModify() {
	data, _ := json.Marshal(g.SkillsArray)
	g.Skills = string(data)
}

// BeforeCreate 钩子函数，在插入数据之前调用
func (g *General) BeforeCreate(tx *gorm.DB) (err error) {
	g.beforeModify()
	return nil
}

// BeforeUpdate 钩子函数，在更新数据之前调用
func (g *General) BeforeUpdate(tx *gorm.DB) (err error) {
	g.beforeModify()
	return nil
}
func (g *General) BeforeFind(tx *gorm.DB) (err error) {
	g.beforeModify()
	return nil
}
func (g *General) AfterFind(tx *gorm.DB) (err error) {
	// 查询后执行的逻辑，比如将 `Skills` 转换为 `SkillsArray`
	err = json.Unmarshal([]byte(g.Skills), &g.SkillsArray)
	if err != nil {
		return err
	}
	return nil
}

/* 推送同步 begin */
func (g *General) IsCellView() bool {
	return false
}

func (g *General) IsCanView(rid, x, y int) bool {
	return false
}

func (g *General) BelongToRId() []int {
	return []int{g.RId}
}

func (g *General) PushMsgName() string {
	return "general.push"
}

func (g *General) Position() (int, int) {
	return -1, -1
}

func (g *General) TPosition() (int, int) {
	return -1, -1
}

func (g *General) Push() {
	net.Mgr.Push(g)
}
