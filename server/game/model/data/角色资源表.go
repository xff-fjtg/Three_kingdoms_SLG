package data

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/game/model"
	"log"
)

var RoleResDao = &roleResDao{
	rrchan: make(chan *RoleRes, 100),
}

type roleResDao struct {
	rrchan chan *RoleRes
}

func init() {
	go RoleResDao.running()

}

type RoleRes struct {
	Id     int `gorm:"column:id;primaryKey;autoIncrement"`
	RId    int `gorm:"column:rid"`
	Wood   int `gorm:"column:wood"`
	Iron   int `gorm:"column:iron"`
	Stone  int `gorm:"column:stone"`
	Grain  int `gorm:"column:grain"`
	Gold   int `gorm:"column:gold"`
	Decree int `gorm:"column:decree"` //令牌
}
type Yield struct { //产量
	Wood  int
	Iron  int
	Stone int
	Grain int
	Gold  int
}

func (r *RoleRes) TableName() string {
	return "role_res"
}

func (r *RoleRes) ToModel() interface{} {
	p := model.RoleRes{}
	p.Gold = r.Gold
	p.Grain = r.Grain
	p.Stone = r.Stone
	p.Iron = r.Iron
	p.Wood = r.Wood
	p.Decree = r.Decree
	yield := GetYield(r.RId)
	p.GoldYield = yield.Gold
	p.GrainYield = yield.Grain
	p.StoneYield = yield.Stone
	p.IronYield = yield.Iron
	p.WoodYield = yield.Wood
	p.DepotCapacity = 10000
	return p
}

func (r *roleResDao) running() {
	for {
		select {
		case rr := <-r.rrchan:
			if rr.Id > 0 {
				// 假设 rr 是 *data.RoleResource 类型，并且已经有 ID 值
				if err := global.DB.Model(rr).Select("wood", "iron", "stone", "grain", "gold", "decree").Where("id = ?", rr.Id).Updates(&rr).Error; err != nil {
					log.Println("roleResDao update error", err)
				}
			}
		}
	}
}
func (r *RoleRes) SyncExecute() {
	RoleResDao.rrchan <- r
	r.Push()
}

/* 推送同步 begin */
func (r *RoleRes) IsCellView() bool {
	return false
}

func (r *RoleRes) IsCanView(rid, x, y int) bool {
	return false
}

func (r *RoleRes) BelongToRId() []int {
	return []int{r.RId}
}

func (r *RoleRes) PushMsgName() string {
	return "roleRes.push"
}

func (r *RoleRes) Position() (int, int) {
	return -1, -1
}

func (r *RoleRes) TPosition() (int, int) {
	return -1, -1
}

func (r *RoleRes) Push() {
	net.Mgr.Push(r)
}
