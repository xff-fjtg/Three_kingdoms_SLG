package logic

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/gameConfig/general"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"encoding/json"
	"log"
	"math/rand"
	"time"
)

type roleWarService struct {
}

var RoleWar = &roleWarService{}

func (r *roleWarService) GetWarReports(rid int) ([]model.WarReport, error) {
	roleWar := make([]data.WarReport, 0)
	err := global.DB.Where("a_rid = ? or d_rid = ?", rid, rid).Find(&roleWar).Error //攻防和收方都是你 你都可以看
	if err != nil {
		log.Println("search roleWar fail")
		return nil, common.New(utils.DBError, "search roleArmy fail")
	}
	modelWar := make([]model.WarReport, 0)
	for _, v := range roleWar {
		modelWar = append(modelWar, v.ToModel().(model.WarReport))
	}
	return modelWar, nil
}

// 判断免战
func IsWarFree(x int, y int) bool {
	//判断是不是城池
	rb, ok := RoleBuild.PositionBuild(x, y)
	if ok {
		return rb.IsWarFree()
	}

	rc, ok := RoleCity.PositionCity(x, y)
	if ok {
		rr, err := RoleAttrService.Get(rc.RId)
		if err != nil {
			log.Println(err)
			return false
		}
		if rr.ParentId > 0 { //已经沦陷不能攻击

			return rc.IsWarFree()
		}
	}
	return false
}

func IsCanDefend(x int, y int, rid int) bool {
	unionId := data.GetUnion(rid)
	rb, ok := RoleBuild.PositionBuild(x, y)
	if ok {
		//拿到当前建筑的联盟id在判断
		toUnionId := data.GetUnion(rb.RId)
		parentId := data.GetParentId(rb.RId)
		if rb.RId == rid {
			return true
		}
		if unionId == toUnionId || unionId == parentId {
			return true
		}
	}
	//查询对应城池
	rc, ok := RoleCity.PositionCity(x, y)
	if ok {
		toUnionId := data.GetUnion(rc.RId)
		parentId := data.GetParentId(rc.RId)
		if rc.RId == rid {
			return true
		}
		if unionId == toUnionId || unionId == parentId {
			return true
		}
	}
	return false
}
func NewEmptyWar(attack *data.Army) *data.WarReport {
	//战报处理
	pArmy := attack.ToModel().(model.Army)
	begArmy, _ := json.Marshal(pArmy)

	//武将战斗前
	begGeneral := make([][]int, 0)
	for _, g := range attack.Gens {
		if g != nil {
			pg := g.ToModel().(model.General)
			begGeneral = append(begGeneral, pg.ToArray())
		}
	}
	begGeneralData, _ := json.Marshal(begGeneral)

	wr := &data.WarReport{X: attack.ToX, Y: attack.ToY, AttackRid: attack.RId,
		AttackIsRead: false, DefenseIsRead: true, DefenseRid: 0,
		BegAttackArmy: string(begArmy), BegDefenseArmy: "",
		EndAttackArmy: string(begArmy), EndDefenseArmy: "",
		BegAttackGeneral:  string(begGeneralData),
		EndAttackGeneral:  string(begGeneralData),
		BegDefenseGeneral: "",
		EndDefenseGeneral: "",
		Rounds:            "",
		Result:            0,
		CTime:             time.Now(),
	}
	return wr
}

// 战斗位置的属性
type armyPosition struct {
	General  *data.General
	Soldiers int //兵力
	Force    int //武力
	Strategy int //策略
	Defense  int //防御
	Speed    int //速度
	Destroy  int //破坏
	Arms     int //兵种
	Position int //位置
}

type Battle struct {
	AId   int `json:"a_id"`   //本回合发起攻击的武将id
	DId   int `json:"d_id"`   //本回合防御方的武将id
	ALoss int `json:"a_loss"` //本回合攻击方损失的兵力
	DLoss int `json:"d_loss"` //本回合防守方损失的兵力
}

func (b *Battle) to() []int {
	r := make([]int, 0)
	r = append(r, b.AId)
	r = append(r, b.DId)
	r = append(r, b.ALoss)
	r = append(r, b.DLoss)
	return r
}

// 最大回合数
const maxRound = 10

type warRound struct {
	Battle [][]int `json:"b"` //每个回合发生的事
}

type WarResult struct {
	Round  []*warRound //每个回合
	Result int         //0失败，1平，2胜利
}

type ArmyWar struct {
	Attack     *data.Army
	Defense    *data.Army
	AttackPos  []*armyPosition
	DefensePos []*armyPosition
}

func (w *ArmyWar) Init() {
	//城内设施加成
	attackAdds := []int{0, 0, 0, 0}
	//根据设施 加攻击 比如什么什么可以加攻击
	if w.Attack.CityId > 0 {
		attackAdds = CityFacilityService.GetAdditions(w.Attack.CityId,
			gameConfig.TypeForce,
			gameConfig.TypeDefense,
			gameConfig.TypeSpeed,
			gameConfig.TypeStrategy)
	}
	//根据设施 加防御 比如什么什么可以加防御
	defenseAdds := []int{0, 0, 0, 0}
	if w.Defense.CityId > 0 {
		defenseAdds = CityFacilityService.GetAdditions(w.Defense.CityId,
			gameConfig.TypeForce,
			gameConfig.TypeDefense,
			gameConfig.TypeSpeed,
			gameConfig.TypeStrategy)
	}

	//TODO 阵营加成

	w.AttackPos = make([]*armyPosition, 0)
	w.DefensePos = make([]*armyPosition, 0)
	//计算一下两边的属性
	for i, g := range w.Attack.Gens {
		if g == nil {
			w.AttackPos = append(w.AttackPos, nil)
		} else {
			pos := &armyPosition{
				General:  g,
				Soldiers: w.Attack.SoldierArray[i],
				Force:    g.GetForce() + attackAdds[0],
				Defense:  g.GetDefense() + attackAdds[1],
				Speed:    g.GetSpeed() + attackAdds[2],
				Strategy: g.GetStrategy() + attackAdds[3],
				Destroy:  g.GetDestroy(),
				Arms:     g.CurArms,
				Position: i,
			}
			w.AttackPos = append(w.AttackPos, pos)
		}
	}

	for i, g := range w.Defense.Gens {
		if g == nil {
			w.DefensePos = append(w.DefensePos, nil)
		} else {
			pos := &armyPosition{
				General:  g,
				Soldiers: w.Defense.SoldierArray[i],
				Force:    g.GetForce() + defenseAdds[0],
				Defense:  g.GetDefense() + defenseAdds[1],
				Speed:    g.GetSpeed() + defenseAdds[2],
				Strategy: g.GetStrategy() + defenseAdds[3],
				Destroy:  g.GetDestroy(),
				Arms:     g.CurArms,
				Position: i,
			}
			w.DefensePos = append(w.DefensePos, pos)
		}
	}
}

func (w *ArmyWar) Battle() []*warRound {
	//随机出手 根据攻击和防御 扣减 士兵
	//结束条件 士兵 主将 为0 或者 达到最大回合数
	rounds := make([]*warRound, 0)
	cur := 0
	for true {
		r, isEnd := w.Round()
		rounds = append(rounds, r)
		cur += 1
		//结束或者达到最大回合 就结束
		if cur >= maxRound || isEnd {
			break
		}
	}
	//战斗完要更新士兵什么的
	for i := 0; i < 3; i++ {
		if w.AttackPos[i] != nil {
			w.Attack.SoldierArray[i] = w.AttackPos[i].Soldiers
		}
		if w.DefensePos[i] != nil {
			w.Defense.SoldierArray[i] = w.DefensePos[i].Soldiers
		}
	}

	return rounds
}

func (w *ArmyWar) Round() (*warRound, bool) {
	war := &warRound{}
	n := rand.Intn(10)
	attack := w.AttackPos
	defense := w.DefensePos

	isEnd := false
	//随机先手
	if n%2 == 0 {
		attack = w.DefensePos
		defense = w.AttackPos
	}

	for _, att := range attack {

		////////攻击方begin//////////
		if att == nil || att.Soldiers == 0 {
			continue
		}
		//看谁来防守 三个防守的来一个
		def, _ := w.randArmyPosition(defense)
		if def == nil {
			isEnd = true
			goto end
		}
		//计算兵种信息
		attHarmRatio := general.GenArms.GetHarmRatio(att.Arms, def.Arms)
		//计算伤害
		attHarm := float64(utils.AbsInt(att.Force-def.Defense)*att.Soldiers) * attHarmRatio * 0.0005
		if att.Force < def.Defense {
			//伤害减免
			attHarm = attHarm * 0.1
		}
		attKill := int(attHarm)
		//自己伤害不能超过对方兵力
		attKill = utils.MinInt(attKill, def.Soldiers)
		def.Soldiers -= attKill
		att.General.Exp += attKill * 5

		//大营干死了，直接结束
		if def.Position == 0 && def.Soldiers == 0 {
			b := Battle{AId: att.General.Id, ALoss: 0, DId: def.General.Id, DLoss: attKill}
			war.Battle = append(war.Battle, b.to())
			isEnd = true
			goto end
		}
		////////攻击方end//////////

		////////防守方begin//////////
		if def.Soldiers == 0 || att.Soldiers == 0 {
			continue
		}

		defHarmRatio := general.GenArms.GetHarmRatio(def.Arms, att.Arms)
		defHarm := float64(utils.AbsInt(def.Force-att.Defense)*def.Soldiers) * defHarmRatio * 0.0005
		defKill := int(defHarm)
		if def.Force < att.Defense {
			//伤害减免
			defHarm = defHarm * 0.1
		}
		defKill = utils.MinInt(defKill, att.Soldiers)
		att.Soldiers -= defKill
		def.General.Exp += defKill * 5
		//记录一下
		b := Battle{AId: att.General.Id, ALoss: defKill, DId: def.General.Id, DLoss: attKill}
		war.Battle = append(war.Battle, b.to())

		//大营干死了，直接结束
		if att.Position == 0 && att.Soldiers == 0 {
			isEnd = true
			goto end
		}
		////////防守方end//////////

	}

end:
	return war, isEnd
}

func (w *ArmyWar) randArmyPosition(pos []*armyPosition) (*armyPosition, int) {
	isEmpty := true
	for _, v := range pos {
		if v != nil && v.Soldiers != 0 {
			isEmpty = false
			break
		}
	}

	if isEmpty {
		return nil, -1
	}
	//看看谁来接受伤害
	for true {
		r := rand.Intn(100)
		index := r % len(pos)
		if pos[index] != nil && pos[index].Soldiers != 0 {
			return pos[index], index
		}
	}

	return nil, -1
}
