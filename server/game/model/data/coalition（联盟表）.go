package data

import (
	"Three_kingdoms_SLG/net"
	"encoding/json"
	"gorm.io/gorm"
	"log"

	//"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model"
	"time"
)

type Coalition struct {
	Id           int       `gorm:"primaryKey;autoIncrement;column:id"`
	Name         string    `gorm:"column:name"`
	Members      string    `gorm:"column:members"`
	MemberArray  []int     `gorm:"-"`
	CreateID     int       `gorm:"column:create_id"`
	Chairman     int       `gorm:"column:chairman"`
	ViceChairman int       `gorm:"column:vice_chairman"`
	Notice       string    `gorm:"column:notice"`
	State        int8      `gorm:"column:state"`
	Ctime        time.Time `gorm:"column:ctime"`
}

func (c *Coalition) TableName() string {
	return "coalition" // 如果数据库中的表名是这个
}

func (c *Coalition) ToModel() interface{} {
	u := model.Union{}
	u.Name = c.Name
	u.Notice = c.Notice
	u.Id = c.Id
	u.Cnt = c.Cnt()
	return u
}

func (c *Coalition) Cnt() int {
	return len(c.MemberArray)
}

type CoalitionApply struct {
	ID      int       `gorm:"primaryKey;autoIncrement;column:id"`
	UnionID int       `gorm:"column:union_id"`
	RId     int       `gorm:"column:rid"`
	State   int8      `gorm:"column:state"`
	Ctime   time.Time `gorm:"column:ctime"`
}

func (d *CoalitionApply) TableName() string {
	return "coalition_apply" // 如果数据库中的表名是这个
}

const (
	UnionOpCreate    = 0 //创建
	UnionOpDismiss   = 1 //解散
	UnionOpJoin      = 2 //加入
	UnionOpExit      = 3 //退出
	UnionOpKick      = 4 //踢出
	UnionOpAppoint   = 5 //任命
	UnionOpAbdicate  = 6 //禅让
	UnionOpModNotice = 7 //修改公告
)
const (
	UnionDismiss = 0 //解散
	UnionRunning = 1 //运行中
)

//	func (c *Coalition) AfterSet(name string, cell xorm.Cell)  {
//		if name == "members"{
//			if cell != nil{
//				ss, ok := (*cell).([]uint8)
//				if ok {
//					json.Unmarshal(ss, &c.MemberArray)
//				}
//				if c.MemberArray == nil{
//					c.MemberArray = []int{}
//					log.Println("查询联盟后进行数据转换",c.MemberArray)
//				}
//			}
//		}
//	}
func (c *Coalition) AfterFind(tx *gorm.DB) (err error) {
	// 检查 Members 字段是否为空
	if c.Members != "" {
		// 尝试将 Members 字段解析为 MemberArray
		err := json.Unmarshal([]byte(c.Members), &c.MemberArray)
		if err != nil {
			log.Printf("解析 Members 字段时出错: %v", err)
			return err
		}
	}
	// 如果解析后为空数组，则初始化空数组
	if c.MemberArray == nil {
		c.MemberArray = []int{}
		log.Println("查询联盟后进行数据转换", c.MemberArray)
	}
	return nil
}

/* 推送同步 begin */
//联盟的申请
func (c *CoalitionApply) ToModel() interface{} {
	//u := model.Union{}
	panic("implement me")
}
func (c *CoalitionApply) IsCellView() bool {
	return false
}

func (c *CoalitionApply) IsCanView(rid, x, y int) bool {
	return false
}

func (c *CoalitionApply) BelongToRId() []int {
	r := GetMainMembers(c.UnionID)
	return append(r, c.RId)
}

func (c *CoalitionApply) PushMsgName() string {
	return "unionApply.push"
}

func (c *CoalitionApply) Position() (int, int) {
	return -1, -1
}

func (c *CoalitionApply) TPosition() (int, int) {
	return -1, -1
}

func (c *CoalitionApply) Push() {
	net.Mgr.Push(c)
}

//func (c *CoalitionApply) ToProto() interface{} {
//	p := model.ApplyItem{}
//	p.RId = c.RId
//	p.Id = c.Id
//	p.NickName = GetRoleNickName(c.RId)
//	return p
//}

func (c *CoalitionApply) SyncExecute() {
	c.Push()
}
