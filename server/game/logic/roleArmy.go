package logic

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/gameConfig/general"
	"Three_kingdoms_SLG/server/game/globalSet"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"encoding/json"
	"log"
	"sync"
	"time"
)

type roleArmyService struct {
	//缓存到达时间和军队的映射关系
	arriveArmyChan chan *data.Army
	//更新的操作
	updateArmyChan chan *data.Army
	giveUpChan     chan int
	endTimeArmys   map[int64][]*data.Army
	//驻守的军队 key posId
	stopInPosArmys map[int]map[int]*data.Army
	passBy         sync.RWMutex
	//就是走路的时候有一条线
	passByPosArmys map[int]map[int]*data.Army //玩家路过位置的军队 key:posId,armyId
	sys            *sysArmyService
}

var RoleArmy = &roleArmyService{
	sys:            NewSysArmy(), //系统武将
	updateArmyChan: make(chan *data.Army, 100),
	giveUpChan:     make(chan int, 100),
	arriveArmyChan: make(chan *data.Army, 100),
	endTimeArmys:   make(map[int64][]*data.Army),
	passByPosArmys: make(map[int]map[int]*data.Army),
	stopInPosArmys: make(map[int]map[int]*data.Army),
}

func (r *roleArmyService) Init() {
	//初始化
	go r.check()
	go r.running()
}
func (r *roleArmyService) check() {
	for {
		time.Sleep(time.Millisecond * 200)
		armysMap := r.endTimeArmys
		cur := time.Now().Unix()
		for endTime, armys := range armysMap {
			if endTime <= cur {
				r.Arrve(armys)
				//已经处理完了
				delete(armysMap, endTime)
			}

		}
	}
}
func (r *roleArmyService) running() {
	//初始化
	for {
		select {
		case army := <-r.arriveArmyChan:
			r.exeArrive(army)
		case army := <-r.updateArmyChan:
			r.exeUpdate(army)
		case posId := <-r.giveUpChan:
			r.GiveUp(posId)
		}
	}
	go r.check()
}
func (r *roleArmyService) GetRoleArmy(rid int) ([]model.Army, error) {
	roleArmy := make([]data.Army, 0)
	err := global.DB.Where("rid = ?", rid).Find(&roleArmy).Error
	if err != nil {
		log.Println("search roleArmy fail")
		return nil, common.New(utils.DBError, "search roleArmy fail")
	}
	modelArmy := make([]model.Army, 0)
	for _, v := range roleArmy {
		modelArmy = append(modelArmy, v.ToModel().(model.Army))
	}
	return modelArmy, nil
}
func (r *roleArmyService) GetRoleArmyByCity(rid, cid int) ([]model.Army, error) {
	roleArmy := make([]*data.Army, 0)

	err := global.DB.Where("rid = ? and cityId = ?", rid, cid).Find(&roleArmy).Error
	if err != nil {
		log.Println("search roleArmy fail")
		return nil, common.New(utils.DBError, "search roleArmy fail")
	}
	modelArmy := make([]model.Army, 0)
	for _, v := range roleArmy {
		modelArmy = append(modelArmy, v.ToModel().(model.Army))
	}
	return modelArmy, nil
}

func (a *roleArmyService) ScanBuild(roleId int, req *model.ScanBlockReq) ([]model.Army, error) {
	x := req.X
	y := req.Y
	length := req.Length
	maxX := utils.MinInt(globalSet.MapWidth, x+length-1)
	maxY := utils.MinInt(globalSet.MapHeight, y+length-1)
	out := make([]model.Army, 0)
	if x < 0 || x >= maxX || y < 0 || y >= maxY {
		return out, nil
	}

	a.passBy.RLock()
	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {

			posId := globalSet.ToPosition(i, j)
			armys, ok := a.passByPosArmys[posId]
			if ok {
				//是否在视野范围内，在就添加，不在就不添加
				is := armyIsInView(roleId, i, j)
				if is == false {
					continue
				}
				for _, army := range armys {
					out = append(out, army.ToModel().(model.Army))
				}
			}
		}
	}
	a.passBy.RUnlock()
	return out, nil
}

func (a *roleArmyService) GetOrCreate(rid int, cid int, order int8) (*data.Army, error) {
	//根据城池id 角色id 和order 查找是否有这个军队
	//有就返回 不然就创建
	armys, err := a.GetRoleArmyByCityAndOrder(rid, cid, order)
	if armys.RId != 0 && armys.CityId != 0 && armys.Id != 0 && armys != nil {
		return armys, err
	}

	//需要创建
	army := &data.Army{RId: rid,
		Order:              order,
		CityId:             cid,
		Generals:           `[0,0,0]`,
		Soldiers:           `[0,0,0]`,
		GeneralArray:       []int{0, 0, 0},
		SoldierArray:       []int{0, 0, 0},
		ConscriptCnts:      `[0,0,0]`,
		ConscriptTimes:     `[0,0,0]`,
		ConscriptCntArray:  []int{0, 0, 0},
		ConscriptTimeArray: []int64{0, 0, 0},
	}

	//city, ok := Default.Get(cid)
	//if ok {
	//	army.FromX = city.X
	//	army.FromY = city.Y
	//	army.ToX = city.X
	//	army.ToY = city.Y
	//}
	a.updateGenerals(army)
	if err := global.DB.Create(&army).Error; err != nil {
		log.Println("create army error", err)
		return nil, err
	}
	return army, nil
}

// 查询具体将领信息
func (a *roleArmyService) updateGenerals(armys ...*data.Army) {
	for _, army := range armys {
		army.Gens = make([]*data.General, 0)
		for _, gid := range army.GeneralArray {
			if gid == 0 {
				army.Gens = append(army.Gens, nil)
			} else {
				g, _ := RoleGeneral.GetByGId(gid) //有详情就可以在页面上看到了
				army.Gens = append(army.Gens, g)
			}
		}
	}
}
func (r *roleArmyService) GetRoleArmyByCityAndOrder(rid, cid int, order int8) (*data.Army, error) {
	armys := &data.Army{}
	//armys := make([]*data.Army, 0)
	err := global.DB.Where("rid = ? and cityId = ? and a_order = ? ", rid, cid, order).Find(&armys).Error
	if err != nil {
		log.Println("search roleArmy fail")
		return nil, common.New(utils.DBError, "search roleArmy fail")
	}
	armys.CheckConscript()
	r.updateGenerals(armys)
	return armys, nil
}

func (a *roleArmyService) IsRepeat(rid int, cfgId int) bool {
	armys, err := a.GetDBArmy(rid)
	if err != nil {
		return true
	}
	for _, army := range armys {
		for _, g := range army.Gens {
			if g != nil {
				if g.CfgId == cfgId && g.CityId != 0 {
					return false
				}
			}
		}
	}
	return true
}
func (r *roleArmyService) GetDBArmy(rid int) ([]*data.Army, error) {
	roleArmy := make([]*data.Army, 0)
	err := global.DB.Where("rid = ?", rid).Find(&roleArmy).Error
	if err != nil {
		log.Println("search roleArmy fail")
		return nil, common.New(utils.DBError, "search roleArmy fail")
	}
	for _, v := range roleArmy {
		v.CheckConscript()
		r.updateGenerals(v)
	}
	return roleArmy, nil
}

func (a *roleArmyService) Get(id int) (*data.Army, bool) {
	army := &data.Army{}
	result := global.DB.Where("id=?", id).Find(&army)
	if result.Error != nil {
		log.Println("军队查询出错", result.Error)
		return nil, false
	}
	army.CheckConscript()
	a.updateGenerals(army)
	return army, true
}

func (a *roleArmyService) GetArmy(cid int, order int8) *data.Army {
	army := &data.Army{}
	result := global.DB.Where("cityId=? and a_order=?", cid, order).Find(&army)
	if result.Error != nil {
		log.Println("armyService GetArmy err", result.Error)
		return nil
	}
	//还需要做一步操作  检测一下是否征兵完成
	army.CheckConscript()
	a.updateGenerals(army)
	return army
}

func (a *roleArmyService) PushAction(army *data.Army) {
	if army.Cmd == data.ArmyCmdAttack {
		//更新缓存和数据库
		//army.End.Unix() 到达时间
		_, ok := a.endTimeArmys[army.End.Unix()]
		if !ok {
			a.endTimeArmys[army.End.Unix()] = make([]*data.Army, 0)
		}
		//到达时间缓存起来了
		a.endTimeArmys[army.End.Unix()] = append(a.endTimeArmys[army.End.Unix()], army)
	} else if army.Cmd == data.ArmyCmdBack {
		//更新缓存和数据库
		//army.End.Unix() 到达时间
		_, ok := a.endTimeArmys[army.End.Unix()]
		if !ok {
			a.endTimeArmys[army.End.Unix()] = make([]*data.Army, 0)
		}
		//到达时间缓存起来了
		a.endTimeArmys[army.End.Unix()] = append(a.endTimeArmys[army.End.Unix()], army)
		army.Start = time.Now()
	}

}

func (r *roleArmyService) Arrve(armys []*data.Army) {
	for _, army := range armys {
		r.arriveArmyChan <- army
	}

}

func (r *roleArmyService) exeArrive(army *data.Army) {
	//开启战争
	if army.Cmd == data.ArmyCmdAttack {
		if !IsWarFree(army.ToX, army.ToY) &&
			!IsCanDefend(army.ToX, army.ToY, army.RId) {
			r.newBattle(army)
		} else {
			wr := NewEmptyWar(army)
			wr.SyncExecute()
		}
		army.State = data.ArmyStop
		r.Updata(army)
	} else if army.Cmd == data.ArmyCmdBack {
		//回城成功
		army.ToX = army.FromX
		army.ToY = army.FromY
		army.State = data.ArmyStop
		army.Cmd = data.ArmyCmdIdle
		r.Updata(army)
	}
}

func (a *roleArmyService) newBattle(attackArmy *data.Army) {
	//1.打土地 建筑 或者 2. 和玩家对打
	//查要操作的城池
	city, ok := RoleCity.PositionCity(attackArmy.ToX, attackArmy.ToY)
	if ok {
		//驻守队伍被打 先判断 这个要打的地方 有没有 部队
		posId := globalSet.ToPosition(attackArmy.ToX, attackArmy.ToY)
		//得到驻守军队
		enemys := a.GetStopArmys(posId)
		//城内空闲的队伍被打
		armys := a.GetArmysByCityId(city.CityId)
		for _, enemy := range armys {
			if enemy.IsCanOutWar() {
				enemys = append(enemys, enemy)
			}
		}

		if len(enemys) == 0 {
			//没有军队 直接攻打 扣除耐久
			destory := RoleGeneral.GetDestroy(attackArmy)
			city.DurableChange(-destory)
			city.SyncExecute()
			//生成空战报
			wr := NewEmptyWar(attackArmy)
			wr.Result = 2
			wr.DefenseRid = city.RId //防守方id
			wr.DefenseIsRead = false //是否阅读
			//判断城池耐久是否为0
			checkCityOccupy(wr, attackArmy, city)
			wr.SyncExecute()
		} else { //有军队
			//打仗
			lastWar, warReports := trigger(attackArmy, enemys, true)
			if lastWar.Result > 1 { //成功
				wr := warReports[len(warReports)-1]
				checkCityOccupy(wr, attackArmy, city)
			}
			for _, wr := range warReports {
				wr.SyncExecute()
			}
		}
	} else {
		//打建筑
		executeBuild(attackArmy)
	}
}

func executeBuild(army *data.Army) {
	//根据攻击位置找到建筑
	roleBuild, _ := RoleBuild.PositionBuild(army.ToX, army.ToY)

	posId := globalSet.ToPosition(army.ToX, army.ToY)
	//看看这个地方有没有军队 有就说明是某个玩家的领地
	posArmys := RoleArmy.GetStopArmys(posId)
	isRoleEnemy := len(posArmys) != 0
	var enemys []*data.Army
	//没有就随机生成一个npc军队
	if isRoleEnemy == false {
		enemys = RoleArmy.sys.GetArmy(army.ToX, army.ToY)
	} else {
		for _, v := range posArmys {
			enemys = append(enemys, v)
		}
	}
	//打仗
	lastWar, warReports := trigger(army, enemys, isRoleEnemy)
	//赢了
	if lastWar.Result > 1 {
		if roleBuild != nil {
			destory := RoleGeneral.GetDestroy(army)
			wr := warReports[len(warReports)-1]
			wr.DestroyDurable = utils.MinInt(destory, roleBuild.CurDurable)
			roleBuild.CurDurable = utils.MaxInt(0, roleBuild.CurDurable-destory)
			if roleBuild.CurDurable == 0 {
				//攻占了玩家的领地
				bLimit := gameConfig.Base.Role.BuildLimit
				//没有超过限额就可以占领
				if bLimit > RoleBuild.BuildCnt(army.RId) {
					//
					wr.Occupy = 1
					RoleBuild.RemoveFromRole(roleBuild)
					RoleBuild.AddBuild(army.RId, army.ToX, army.ToY)
					OccupyRoleBuild(army.RId, army.ToX, army.ToY)
				} else {
					wr.Occupy = 0
				}
			} else {
				wr.Occupy = 0
			}

		} else {
			//占领系统领地
			wr := warReports[len(warReports)-1]
			blimit := gameConfig.Base.Role.BuildLimit
			if blimit > RoleBuild.BuildCnt(army.RId) {
				//占领
				OccupySystemBuild(army.RId, army.ToX, army.ToY)
				wr.DestroyDurable = 10000
				wr.Occupy = 1
			} else {
				wr.Occupy = 0
			}
			RoleArmy.sys.DelArmy(army.ToX, army.ToY)
		}
	}

	//领地发生变化
	if newRoleBuild, ok := RoleBuild.PositionBuild(army.ToX, army.ToY); ok {
		newRoleBuild.SyncExecute()
	}

	for _, wr := range warReports {
		wr.SyncExecute()
	}
}

func OccupySystemBuild(rid int, x int, y int) {
	if _, ok := RoleBuild.PositionBuild(x, y); ok {
		return
	}
	//判断能不能被占领
	if gameConfig.MapRes.IsCanBuild(x, y) {
		rb, ok := RoleBuild.AddBuild(rid, x, y)
		if ok {
			rb.OccupyTime = time.Now()
			rb.SyncExecute()
		}
	}
}

func OccupyRoleBuild(rid int, x int, y int) {
	if b, ok := RoleBuild.PositionBuild(x, y); ok {

		b.CurDurable = b.MaxDurable
		b.OccupyTime = time.Now()
		b.RId = rid
		b.SyncExecute()
	}
}

func trigger(army *data.Army, enemys []*data.Army, isRoleEnemy bool) (*WarResult, []*data.WarReport) {
	//拿到位置
	posId := globalSet.ToPosition(army.ToX, army.ToY)
	warReports := make([]*data.WarReport, 0)
	var lastWar *WarResult = nil
	//一个敌军 一个战报
	for _, enemy := range enemys {
		//战报处理
		//先把敌我双方的部队拿出来
		pArmy := army.ToModel().(model.Army)
		pEnemy := enemy.ToModel().(model.Army)

		begArmy1, _ := json.Marshal(pArmy)
		begArmy2, _ := json.Marshal(pEnemy)

		//武将战斗前 敌我的武将拿出来 生成战前战报
		begGeneral1 := make([][]int, 0)
		for _, g := range army.Gens {
			if g != nil {
				pg := g.ToModel().(model.General)
				begGeneral1 = append(begGeneral1, pg.ToArray())
			}
		}
		begGeneralData1, _ := json.Marshal(begGeneral1)

		begGeneral2 := make([][]int, 0)
		for _, g := range enemy.Gens {
			if g != nil {
				pg := g.ToModel().(model.General)
				begGeneral2 = append(begGeneral2, pg.ToArray())
			}
		}
		begGeneralData2, _ := json.Marshal(begGeneral2)
		//去战斗
		lastWar = NewWar(army, enemy)

		//武将战斗后 处理结果
		endGeneral1 := make([][]int, 0)
		for _, g := range army.Gens {
			if g != nil {
				pg := g.ToModel().(model.General)
				endGeneral1 = append(endGeneral1, pg.ToArray())
				//能不能升级
				level, exp := general.GeneralBasic.ExpToLevel(g.Exp)
				g.Level = level
				g.Exp = exp
				g.SyncExecute()
			}
		}
		endGeneralData1, _ := json.Marshal(endGeneral1)

		endGeneral2 := make([][]int, 0)
		for _, g := range enemy.Gens {
			if g != nil {
				pg := g.ToModel().(model.General)
				endGeneral2 = append(endGeneral2, pg.ToArray())
				level, exp := general.GeneralBasic.ExpToLevel(g.Exp)
				g.Level = level
				g.Exp = exp
				g.SyncExecute()
			}
		}
		endGeneralData2, _ := json.Marshal(endGeneral2)

		pArmy = army.ToModel().(model.Army)
		pEnemy = enemy.ToModel().(model.Army)
		endArmy1, _ := json.Marshal(pArmy)
		endArmy2, _ := json.Marshal(pEnemy)

		rounds, _ := json.Marshal(lastWar.Round)
		//生成战报
		wr := &data.WarReport{X: army.ToX, Y: army.ToY, AttackRid: army.RId,
			AttackIsRead: false, DefenseIsRead: false, DefenseRid: enemy.RId,
			BegAttackArmy: string(begArmy1), BegDefenseArmy: string(begArmy2),
			EndAttackArmy: string(endArmy1), EndDefenseArmy: string(endArmy2),
			BegAttackGeneral:  string(begGeneralData1),
			BegDefenseGeneral: string(begGeneralData2),
			EndAttackGeneral:  string(endGeneralData1),
			EndDefenseGeneral: string(endGeneralData2),
			Rounds:            string(rounds),
			Result:            lastWar.Result,
			CTime:             time.Now(),
		}

		warReports = append(warReports, wr)
		enemy.ToSoldier()
		enemy.ToGeneral()
		//是否有玩家的军队
		if isRoleEnemy {
			if lastWar.Result > 1 {
				if isRoleEnemy {
					//被打赢了 占领的军队要删除
					RoleArmy.deleteStopArmy(posId)
				}
				//失败了了回去
				RoleArmy.ArmyBack(enemy)
			}
			enemy.SyncExecute()
		} else {
			wr.DefenseIsRead = true
		}
	}
	army.SyncExecute()
	return lastWar, warReports
}

func NewWar(attack *data.Army, defense *data.Army) *WarResult {
	w := ArmyWar{Attack: attack, Defense: defense}
	w.Init()
	wars := w.Battle()

	result := &WarResult{Round: wars}
	if w.AttackPos[0].Soldiers == 0 {
		result.Result = 0 //自己士兵没了 失败
	} else if w.DefensePos[0] != nil && w.DefensePos[0].Soldiers != 0 {
		result.Result = 1 //平局
	} else {
		result.Result = 2
	}

	return result
}

func (a *roleArmyService) GetStopArmys(posId int) []*data.Army {
	//玩家停留位置的军队
	ret := make([]*data.Army, 0)
	armys, ok := a.stopInPosArmys[posId]
	if ok {
		for _, army := range armys {
			ret = append(ret, army)
		}
	}
	return ret
}

func (r *roleArmyService) GetArmysByCityId(cid int) []*data.Army {
	//army := &data.Army{}
	armys := make([]*data.Army, 0)
	result := global.DB.Where("cityId=? ", cid).Find(&armys)
	if result.Error != nil {
		log.Println("armyService GetArmy err", result.Error)
		return nil
	}
	for _, ar := range armys {
		//还需要做一步操作  检测一下是否征兵完成
		ar.CheckConscript()
		//army.CheckConscript()
		r.updateGenerals(ar)
	}

	return armys
}

func (r *roleArmyService) deleteStopArmy(posId int) {
	delete(r.stopInPosArmys, posId)
}

func (r *roleArmyService) ArmyBack(army *data.Army) {
	army.ClearConscript()

	army.State = data.ArmyRunning
	army.Cmd = data.ArmyCmdBack

	//清除掉之前存的时间
	t := army.End.Unix()
	if actions, ok := r.endTimeArmys[t]; ok {
		for i, v := range actions {
			if v.Id == army.Id {
				actions = append(actions[:i], actions[i+1:]...)
				r.endTimeArmys[t] = actions
				break
			}
		}
	}
	army.Start = time.Now()
	army.End = time.Now().Add(time.Second * 10)
	r.PushAction(army)
}
func armyIsInView(rid, x, y int) bool {
	//简单点 先设为true
	return true
}

// 占领
func checkCityOccupy(wr *data.WarReport, attackArmy *data.Army, city *data.MapRoleCity) {
	destory := RoleGeneral.GetDestroy(attackArmy)
	wr.DestroyDurable = utils.MinInt(destory, city.CurDurable)
	city.DurableChange(-destory)

	if city.CurDurable <= 0 {
		aAttr, _ := RoleAttrService.Get(attackArmy.RId)
		if aAttr.UnionId != 0 {
			//有联盟才能俘虏玩家
			wr.Occupy = 1 //攻占
			dAttr, _ := RoleAttrService.Get(city.RId)
			if dAttr != nil {
				dAttr.ParentId = aAttr.UnionId
				//CoalitionService.PutChild(aAttr.UnionId, city.RId)
				dAttr.SyncExecute()
			}
			city.OccupyTime = time.Now()
		} else {
			wr.Occupy = 0
		}
	} else {
		wr.Occupy = 0
	}
	city.SyncExecute()
}

func (r *roleArmyService) Updata(army *data.Army) {
	r.updateArmyChan <- army
}

func (r *roleArmyService) exeUpdate(army *data.Army) {
	army.SyncExecute()
	if army.Cmd == data.ArmyCmdBack {
		posId := globalSet.ToPosition(army.ToX, army.ToY)
		//当前位置的部队清理掉
		armys, ok := r.stopInPosArmys[posId]
		if ok {
			delete(armys, army.Id)
			r.stopInPosArmys[posId] = armys
		}
	}
}

func (a *roleArmyService) GiveUp(posId int) {
	//该位置驻守的军队需要返回
	armys, ok := a.stopInPosArmys[posId]
	if ok {
		for _, army := range armys {
			a.ArmyBack(army)
		}
		delete(a.stopInPosArmys, posId)
	}

}

func (r *roleArmyService) GiveUpPosId(posId int) {
	r.giveUpChan <- posId
}
