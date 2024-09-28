package data

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/game/globalSet"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/utils"

	//"Three_kingdoms_SLG/server/game/model"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"time"
)

const (
	ArmyCmdIdle        = 0 //空闲
	ArmyCmdAttack      = 1 //攻击
	ArmyCmdDefend      = 2 //驻守
	ArmyCmdReclamation = 3 //屯垦
	ArmyCmdBack        = 4 //撤退
	ArmyCmdConscript   = 5 //征兵
	ArmyCmdTransfer    = 6 //调动
)

const (
	ArmyStop    = 0
	ArmyRunning = 1
)

// 军队
type Army struct {
	Id                 int        `gorm:"column:id;primaryKey;autoIncrement"`
	RId                int        `gorm:"column:rid"`
	CityId             int        `gorm:"column:cityId"`
	Order              int8       `gorm:"column:a_order"`
	Generals           string     `gorm:"column:generals"`
	Soldiers           string     `gorm:"column:soldiers"`
	ConscriptTimes     string     `gorm:"column:conscript_times"` // 征兵结束时间，json数组
	ConscriptCnts      string     `gorm:"column:conscript_cnts"`  // 征兵数量，json数组
	Cmd                int8       `gorm:"column:cmd"`
	FromX              int        `gorm:"column:from_x"`
	FromY              int        `gorm:"column:from_y"`
	ToX                int        `gorm:"column:to_x"`
	ToY                int        `gorm:"column:to_y"`
	Start              time.Time  `gorm:"-"`
	End                time.Time  `gorm:"-"`
	State              int8       `gorm:"-"` // 状态:0:running,1:stop
	GeneralArray       []int      `gorm:"-"`
	SoldierArray       []int      `gorm:"-"`
	ConscriptTimeArray []int64    `gorm:"-"`
	ConscriptCntArray  []int      `gorm:"-"`
	Gens               []*General `gorm:"-"`
	CellX              int        `gorm:"-"`
	CellY              int        `gorm:"-"`
}

// 执行update之前的操作
func (a *Army) BeforeUpdate(tx *gorm.DB) (err error) {
	a.beforeModify()
	return nil
}

func (a *Army) beforeModify() {
	data, _ := json.Marshal(a.GeneralArray)
	a.Generals = string(data)

	data, _ = json.Marshal(a.SoldierArray)
	a.Soldiers = string(data)

	data, _ = json.Marshal(a.ConscriptTimeArray)
	a.ConscriptTimes = string(data)

	data, _ = json.Marshal(a.ConscriptCntArray)
	a.ConscriptCnts = string(data)
}

// 执行insert之前的操作
func (a *Army) BeforeCreate(tx *gorm.DB) (err error) {
	a.beforeModify()
	return nil
}
func (a *Army) TableName() string {
	return "army"
}

func (a *Army) ToModel() interface{} {
	p := model.Army{}
	p.CityId = a.CityId
	p.Id = a.Id
	p.UnionId = GetUnion(a.RId)
	p.Order = a.Order
	p.Generals = a.GeneralArray
	p.Soldiers = a.SoldierArray
	p.ConTimes = a.ConscriptTimeArray
	p.ConCnts = a.ConscriptCntArray
	p.Cmd = a.Cmd
	p.State = a.State
	p.FromX = a.FromX
	p.FromY = a.FromY
	p.ToX = a.ToX
	p.ToY = a.ToY
	p.Start = a.Start.Unix()
	p.End = a.End.Unix()
	return p
}
func (a *Army) AfterFind(tx *gorm.DB) (err error) {
	// Parse "generals"
	a.GeneralArray = []int{0, 0, 0}
	if len(a.Generals) > 0 {
		err = json.Unmarshal([]byte(a.Generals), &a.GeneralArray)
		if err != nil {
			return err
		}
		fmt.Println(a.GeneralArray)
	}

	// Parse "soldiers"
	a.SoldierArray = []int{0, 0, 0}
	if len(a.Soldiers) > 0 {
		err = json.Unmarshal([]byte(a.Soldiers), &a.SoldierArray)
		if err != nil {
			return err
		}
		fmt.Println(a.SoldierArray)
	}

	// Parse "conscript_times"
	a.ConscriptTimeArray = []int64{0, 0, 0}
	if len(a.ConscriptTimes) > 0 {
		err = json.Unmarshal([]byte(a.ConscriptTimes), &a.ConscriptTimeArray)
		if err != nil {
			return err
		}
		fmt.Println(a.ConscriptTimeArray)
	}

	// Parse "conscript_cnts"
	a.ConscriptCntArray = []int{0, 0, 0}
	if len(a.ConscriptCnts) > 0 {
		err = json.Unmarshal([]byte(a.ConscriptCnts), &a.ConscriptCntArray)
		if err != nil {
			return err
		}
		fmt.Println(a.ConscriptCntArray)
	}
	//
	data, _ := json.Marshal(a.GeneralArray)
	a.Generals = string(data)

	data, _ = json.Marshal(a.SoldierArray)
	a.Soldiers = string(data)

	data, _ = json.Marshal(a.ConscriptTimeArray)
	a.ConscriptTimes = string(data)

	data, _ = json.Marshal(a.ConscriptCntArray)
	a.ConscriptCnts = string(data)
	return nil
}

// pos 0-2
func (a *Army) PositionCanModify(pos int) bool {
	if pos >= 3 || pos < 0 {
		return false
	}

	if a.Cmd == ArmyCmdIdle {
		return true
	} else if a.Cmd == ArmyCmdConscript {
		endTime := a.ConscriptTimeArray[pos]
		return endTime == 0
	} else {
		return false
	}
}

func (a *Army) SyncExecute() {
	ArmyDao.aChan <- a
	//同步的时候就把军队实时详情推送出去了
	a.Push()
	a.CellX, a.CellY = a.Position()
}

func (a *Army) CheckConscript() {
	if a.Cmd == ArmyCmdConscript { //正在征兵状态
		curTime := time.Now().Unix()
		finish := true
		for i, endTime := range a.ConscriptTimeArray {
			if endTime > 0 {
				if endTime <= curTime { //征兵完成
					a.SoldierArray[i] += a.ConscriptCntArray[i]
					a.ConscriptCntArray[i] = 0
					a.ConscriptTimeArray[i] = 0
				} else {
					finish = false
				}
			}
		}

		if finish { //设置为征兵完成
			a.Cmd = ArmyCmdIdle
		}
	}
}

func (a *Army) IsCanOutWar() bool {
	//空闲状态
	return a.Gens != nil && a.Cmd == ArmyCmdIdle
}

func (a *Army) IsIdle() bool {
	return a.Cmd == ArmyCmdIdle
}

func (a *Army) ToSoldier() {
	if a.SoldierArray != nil {
		data, _ := json.Marshal(a.SoldierArray)
		a.Soldiers = string(data)
	}
}

func (a *Army) ToGeneral() {
	if a.GeneralArray != nil {
		data, _ := json.Marshal(a.GeneralArray)
		a.Generals = string(data)
	}
}

var ArmyDao = &armyDao{
	aChan: make(chan *Army, 100),
}

type armyDao struct {
	aChan chan *Army
}

func (a *armyDao) running() {
	for {
		select {
		case army := <-a.aChan:
			if army.Id > 0 {
				global.DB.Model(army).Select(
					"soldiers", "generals", "conscript_times",
					"conscript_cnts", "cmd", "from_x", "from_y", "to_x",
					"to_y", "start", "end",
				).Where("id = ?", army.Id).Updates(&army)
			}
		}
	}
}

func init() {
	go ArmyDao.running()
}

// 下面是服务端推送数据给前端
func (a *Army) IsCellView() bool {
	return true
}
func (a *Army) IsCanView(rid, x, y int) bool {
	return true
}
func (a *Army) BelongToRId() []int {
	return []int{a.RId}
}

func (a *Army) PushMsgName() string {
	return "army.push"
}

func (a *Army) Position() (int, int) {
	//时时计算当前位置
	diffTime := a.End.Unix() - a.Start.Unix()
	passTime := time.Now().Unix() - a.Start.Unix()
	rate := float32(passTime) / float32(diffTime)
	x := 0
	y := 0
	if a.Cmd == ArmyCmdBack {
		diffX := a.FromX - a.ToX
		diffY := a.FromY - a.ToY
		x = int(rate*float32(diffX)) + a.ToX
		y = int(rate*float32(diffY)) + a.ToY
	} else {
		diffX := a.ToX - a.FromX
		diffY := a.ToY - a.FromY
		x = int(rate*float32(diffX)) + a.FromX
		y = int(rate*float32(diffY)) + a.FromY
	}

	x = utils.MinInt(utils.MaxInt(x, 0), globalSet.MapWidth)
	y = utils.MinInt(utils.MaxInt(y, 0), globalSet.MapHeight)
	return x, y
}

func (a *Army) TPosition() (int, int) {
	return a.ToX, a.ToY
}

// 消息push出去
func (a *Army) Push() {
	net.Mgr.Push(a)
}

func (a *Army) ClearConscript() {
	if a.Cmd == ArmyCmdConscript {
		for i, _ := range a.ConscriptTimeArray {
			a.ConscriptCntArray[i] = 0
			a.ConscriptTimeArray[i] = 0
		}
		a.Cmd = ArmyCmdIdle
	}
}
