package data

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/game/model"
	"log"
	"time"
)

var RoleAttrDao = &roleAttrDao{
	rachan: make(chan *RoleAttribute, 100),
}

type roleAttrDao struct {
	rachan chan *RoleAttribute
}

func (d *roleAttrDao) running() {
	for {
		select {
		case rr := <-d.rachan:
			if rr.Id > 0 {
				// 假设 rr 是 *data.RoleResource 类型，并且已经有 ID 值
				if err := global.DB.Model(rr).Select("parent_id", "collect_times", "last_collect_time", "pos_tags").Where("id = ?", rr.Id).Updates(rr).Error; err != nil {
					log.Println("roleAttrDao update error", err)
				}
			}
		}
	}
}

type RoleAttribute struct {
	Id              int            `gorm:"column:id;primaryKey;autoIncrement"`
	RId             int            `gorm:"column:rid"`
	UnionId         int            `gorm:"-"`                         // 联盟id，不在数据库中
	ParentId        int            `gorm:"column:parent_id"`          // 上级id（被沦陷）
	CollectTimes    int8           `gorm:"column:collect_times"`      // 征收次数
	LastCollectTime time.Time      `gorm:"column:last_collect_time" ` // 最后征收的时间
	PosTags         string         `gorm:"column:pos_tags"`           // 位置标记
	PosTagArray     []model.PosTag `gorm:"-"`                         // 不在数据库中
}

func init() {
	go RoleAttrDao.running()
}

// 表名自定义（如果需要不同于结构体名）
func (r *RoleAttribute) TableName() string {
	return "role_attribute"
}

func (r *RoleAttribute) SyncExecute() {
	RoleAttrDao.rachan <- r
	r.Push()
}

/* 推送同步 begin */
func (r *RoleAttribute) IsCellView() bool {
	return false
}

func (r *RoleAttribute) IsCanView(rid, x, y int) bool {
	return false
}

func (r *RoleAttribute) BelongToRId() []int {
	return []int{r.RId}
}

func (r *RoleAttribute) PushMsgName() string {
	return "roleAttr.push"
}

func (r *RoleAttribute) ToModel() interface{} {
	return nil
}

func (r *RoleAttribute) Position() (int, int) {
	return -1, -1
}

func (r *RoleAttribute) TPosition() (int, int) {
	return -1, -1
}

func (r *RoleAttribute) Push() {
	net.Mgr.Push(r)
}
