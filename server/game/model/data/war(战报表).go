package data

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/game/model"
	"time"
)

var WarReportDao = &warReportDao{
	wrChan: make(chan *WarReport, 100),
}

type warReportDao struct {
	wrChan chan *WarReport
}

func (w *warReportDao) running() {
	for {
		select {
		case wr := <-w.wrChan:
			if wr.Id <= 0 {
				global.DB.Create(wr)
			} else {
				global.DB.Model(&wr).Updates(&wr)
			}
		}
	}
}
func init() {
	go WarReportDao.running()
}

type WarReport struct {
	Id                int       `gorm:"column:id;primaryKey;autoIncrement"`
	AttackRid         int       `gorm:"column:a_rid"`
	DefenseRid        int       `gorm:"column:d_rid"`
	BegAttackArmy     string    `gorm:"column:b_a_army"`
	BegDefenseArmy    string    `gorm:"column:b_d_army"`
	EndAttackArmy     string    `gorm:"column:e_a_army"`
	EndDefenseArmy    string    `gorm:"column:e_d_army"`
	BegAttackGeneral  string    `gorm:"column:b_a_general"`
	BegDefenseGeneral string    `gorm:"column:b_d_general"`
	EndAttackGeneral  string    `gorm:"column:e_a_general"`
	EndDefenseGeneral string    `gorm:"column:e_d_general"`
	Result            int       `gorm:"column:result"` // 0失败，1打平，2胜利
	Rounds            string    `gorm:"column:rounds"` // 回合
	AttackIsRead      bool      `gorm:"column:a_is_read"`
	DefenseIsRead     bool      `gorm:"column:d_is_read"`
	DestroyDurable    int       `gorm:"column:destroy"`
	Occupy            int       `gorm:"column:occupy"`
	X                 int       `gorm:"column:x"`
	Y                 int       `gorm:"column:y"`
	CTime             time.Time `gorm:"column:ctime"`
}

// TableName allows you to define a custom table name
func (w *WarReport) TableName() string {
	return "war_report"
}

func (w *WarReport) ToModel() interface{} {
	p := model.WarReport{}
	p.CTime = int(w.CTime.UnixNano() / 1e6)
	p.Id = w.Id
	p.AttackRid = w.AttackRid
	p.DefenseRid = w.DefenseRid
	p.BegAttackArmy = w.BegAttackArmy
	p.BegDefenseArmy = w.BegDefenseArmy
	p.EndAttackArmy = w.EndAttackArmy
	p.EndDefenseArmy = w.EndDefenseArmy
	p.BegAttackGeneral = w.BegAttackGeneral
	p.BegDefenseGeneral = w.BegDefenseGeneral
	p.EndAttackGeneral = w.EndAttackGeneral
	p.EndDefenseGeneral = w.EndDefenseGeneral
	p.Result = w.Result
	p.Rounds = w.Rounds
	p.AttackIsRead = w.AttackIsRead
	p.DefenseIsRead = w.DefenseIsRead
	p.DestroyDurable = w.DestroyDurable
	p.Occupy = w.Occupy
	p.X = w.X
	p.X = w.X
	return p
}

func (w *WarReport) SyncExecute() {
	WarReportDao.wrChan <- w
	w.Push()
}

/* 推送同步 begin */
func (w *WarReport) IsCellView() bool {
	return false
}

func (w *WarReport) IsCanView(rid, x, y int) bool {
	return false
}

func (w *WarReport) BelongToRId() []int {
	//战斗有两方 所以要给两方推送
	return []int{w.AttackRid, w.DefenseRid}
}

func (w *WarReport) PushMsgName() string {
	return "warReport.push"
}

func (w *WarReport) Position() (int, int) {
	return w.X, w.Y
}

func (w *WarReport) TPosition() (int, int) {
	return -1, -1
}

func (w *WarReport) Push() {
	net.Mgr.Push(w)
}
