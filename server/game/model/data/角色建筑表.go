package data

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/utils"
	"log"
	"time"
)

const (
	MapBuildSysFortress = 50 //系统要塞
	MapBuildSysCity     = 51 //系统城市
	MapBuildFortress    = 56 //玩家要塞
)

var MapRoleBuildDao = &mapRoleBuildDao{
	rbChan: make(chan *MapRoleBuild, 100),
}

type mapRoleBuildDao struct {
	rbChan chan *MapRoleBuild
}

func (d *mapRoleBuildDao) running() {
	for {
		select {
		case rb := <-d.rbChan:
			{
				if rb.Id > 0 {
					rb.EndTime = time.Now()
					err := global.DB.Model(&rb).Select("rid", "type", "level", "op_level", "cur_durable", "max_durable", "occupy_time", "giveUp_time").Where("id = ?", rb.Id).Updates(rb)

					if err != nil {
						log.Println("create mapRoleBuildDao fail")
					}
				}

			}
		}
	}
}

type MapRoleBuild struct {
	Id         int       `gorm:"column:id;primaryKey;autoIncrement"`
	RId        int       `gorm:"column:rid"`
	Type       int8      `gorm:"column:type"`
	Level      int8      `gorm:"column:level"`
	OPLevel    int8      `gorm:"column:op_level"` // 操作level
	X          int       `gorm:"column:x"`
	Y          int       `gorm:"column:y"`
	Name       string    `gorm:"column:name"`
	Wood       int       `gorm:"-"`
	Iron       int       `gorm:"-"`
	Stone      int       `gorm:"-"`
	Grain      int       `gorm:"-"`
	Defender   int       `gorm:"-"`
	CurDurable int       `gorm:"column:cur_durable"`
	MaxDurable int       `gorm:"column:max_durable"`
	OccupyTime time.Time `gorm:"column:occupy_time"`
	EndTime    time.Time `gorm:"column:end_time"` // 建造或升级完的时间
	GiveUpTime int64     `gorm:"column:giveUp_time"`
}

func (m *MapRoleBuild) TableName() string {
	return "map_role_build"
}

func (m *MapRoleBuild) ToModel() interface{} {
	p := model.MapRoleBuild{}
	p.RNick = "111"
	p.UnionId = 0
	p.UnionName = ""
	p.ParentId = 0
	p.X = m.X
	p.Y = m.Y
	p.Type = m.Type
	p.RId = m.RId
	p.Name = m.Name

	p.OccupyTime = m.OccupyTime.UnixNano() / 1e6
	p.GiveUpTime = m.GiveUpTime * 1000
	p.EndTime = m.EndTime.UnixNano() / 1e6

	p.CurDurable = m.CurDurable
	p.MaxDurable = m.MaxDurable
	p.Defender = m.Defender
	p.Level = m.Level
	p.OPLevel = m.OPLevel
	return p
}

func (m *MapRoleBuild) Init() {
	if cfg := gameConfig.MapBuildConf.BuildConfig(m.Type, m.Level); cfg != nil {
		m.Name = cfg.Name
		m.Level = cfg.Level
		m.Type = cfg.Type
		m.Wood = cfg.Wood
		m.Iron = cfg.Iron
		m.Stone = cfg.Stone
		m.Grain = cfg.Grain
		m.MaxDurable = cfg.Durable
		m.CurDurable = cfg.Durable
		m.Defender = cfg.Defender
	}
}

func (m *MapRoleBuild) IsWarFree() bool {
	var cur = time.Now().Unix()
	//当前时间-占领时间<warfree 说明不能占领
	if cur-m.OccupyTime.Unix() < gameConfig.Base.Build.WarFree {
		return true
	}
	return false
}

func (m *MapRoleBuild) Reset() {
	ok, t, level := MapResTypeLevel(m.X, m.Y)
	if ok {
		if cfg := gameConfig.MapBuildConf.BuildConfig(t, level); cfg != nil {
			m.Name = cfg.Name
			m.Level = cfg.Level
			m.Type = cfg.Type
			m.Wood = cfg.Wood
			m.Iron = cfg.Iron
			m.Stone = cfg.Stone
			m.Grain = cfg.Grain
			m.MaxDurable = cfg.Durable
			m.CurDurable = cfg.Durable
			m.Defender = cfg.Defender
		}
		m.GiveUpTime = 0
		m.RId = 0
		m.EndTime = time.Time{}
		m.OPLevel = m.Level
		m.CurDurable = utils.MinInt(m.MaxDurable, m.CurDurable)
	}

}
func init() {
	go MapRoleBuildDao.running()
}
func (m *MapRoleBuild) SyncExecute() {
	MapRoleBuildDao.rbChan <- m
	//push数据到前端
	m.Push()
}

/* 推送同步 begin */
func (m *MapRoleBuild) IsCellView() bool {
	return true
}

func (m *MapRoleBuild) IsCanView(rid, x, y int) bool {
	return true
}

func (m *MapRoleBuild) BelongToRId() []int {
	return []int{m.RId}
}

func (m *MapRoleBuild) PushMsgName() string {
	return "roleBuild.push"
}

func (m *MapRoleBuild) Position() (int, int) {
	return m.X, m.Y
}

func (m *MapRoleBuild) TPosition() (int, int) {
	return -1, -1
}

func (m *MapRoleBuild) Push() {
	net.Mgr.Push(m)
}

func (m *MapRoleBuild) IsResBuild() bool {
	return m.Grain > 0 || m.Stone > 0 || m.Iron > 0 || m.Wood > 0
}

func (m *MapRoleBuild) IsBusy() bool {
	//自己的level 操作level不一样
	if m.Level != m.OPLevel {
		return true
	} else {
		return false
	}
}

func (m *MapRoleBuild) BuildOrUp(cfg gameConfig.BCLevelCfg) {
	m.Type = cfg.Type
	m.Level = cfg.Level - 1
	m.Name = cfg.Name
	m.OPLevel = cfg.Level
	m.GiveUpTime = 0

	m.Wood = 0
	m.Iron = 0
	m.Stone = 0
	m.Grain = 0
	m.EndTime = time.Now().Add(time.Duration(cfg.Time) * time.Second)
}
