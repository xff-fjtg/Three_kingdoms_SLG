package data

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/model"
	"log"
	"sync"
	"time"
)

var RoleCityDao = &mapRoleCityDao{
	rcChan: make(chan *MapRoleCity, 100),
}

type mapRoleCityDao struct {
	rcChan chan *MapRoleCity
}

func (d *mapRoleCityDao) running() {
	for {
		select {
		case rc := <-d.rcChan:
			if rc.CityId > 0 {
				//更新
				if err := global.DB.Create(&rc).Error; err != nil {
					log.Println("更新玩家城池失败 mapRoleCityDao", err)
				}
			}

		}
	}
}
func init() {
	go RoleCityDao.running()
}

// 玩家进入游戏后，需要加载其拥有的城池，没有就初始化一个做为主城
type MapRoleCity struct {
	mutex      sync.Mutex `gorm:"-"` // 不在数据库中
	CityId     int        `gorm:"column:cityId;primaryKey;autoIncrement"`
	RId        int        `gorm:"column:rid"`
	Name       string     `gorm:"column:name;size:20;not null" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	X          int        `gorm:"column:x"`
	Y          int        `gorm:"column:y"`
	IsMain     int8       `gorm:"column:is_main"`
	CurDurable int        `gorm:"column:cur_durable"` //耐久
	CreatedAt  time.Time  `gorm:"column:created_at"`
	OccupyTime time.Time  `gorm:"column:occupy_time;default:2013-03-15 14:38:09"`
}

// 表名自定义（如果需要不同于结构体名）
func (m *MapRoleCity) TableName() string {
	return "map_role_city"
}

func (m *MapRoleCity) ToModel() interface{} {
	p := model.MapRoleCity{}
	p.X = m.X
	p.Y = m.Y
	p.CityId = m.CityId
	p.UnionId = GetUnion(m.RId)
	p.UnionName = ""
	p.ParentId = 0
	p.MaxDurable = 1000
	p.CurDurable = m.CurDurable
	p.Level = 1
	p.RId = m.RId
	p.Name = m.Name
	p.IsMain = m.IsMain == 1
	p.OccupyTime = m.OccupyTime.UnixNano() / 1e6
	return p
}

func (m *MapRoleCity) IsWarFree() bool {
	var cur = time.Now().Unix()
	//当前时间-占领时间<warfree 说明不能占领
	if cur-m.OccupyTime.Unix() < gameConfig.Base.Build.WarFree {
		return true
	}
	return false
}

func (m *MapRoleCity) DurableChange(change int) {
	t := m.CurDurable + change
	if t < 0 {
		m.CurDurable = 0
	} else {
		//不能大于最大耐久度
		//m.CurDurable = utils.MinInt(GetMaxDurable(m.CityId), t)
	}
}

func (m *MapRoleCity) SyncExecute() {
	RoleCityDao.rcChan <- m
	m.Push()
}

/* 推送同步 begin */
func (m *MapRoleCity) IsCellView() bool {
	return true
}

func (m *MapRoleCity) IsCanView(rid, x, y int) bool {
	return true
}

func (m *MapRoleCity) BelongToRId() []int {
	return []int{m.RId}
}

func (m *MapRoleCity) PushMsgName() string {
	return "roleCity.push"
}

func (m *MapRoleCity) Position() (int, int) {
	return m.X, m.Y
}

func (m *MapRoleCity) TPosition() (int, int) {
	return -1, -1
}
func (m *MapRoleCity) Push() {
	net.Mgr.Push(m)
}
