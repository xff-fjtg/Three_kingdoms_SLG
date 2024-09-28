package controller

import (
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/gameConfig/general"
	"Three_kingdoms_SLG/server/game/globalSet"
	"Three_kingdoms_SLG/server/game/logic"
	"Three_kingdoms_SLG/server/game/middleware"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"log"
	"time"
)

var DefaultArmy = &ArmyController{}

type ArmyController struct {
}

func (a *ArmyController) Router(router *net.Router) {
	g := router.Group("army")
	g.Use(middleware.Log())
	g.AddRouter("myList", a.myList)                               //初始化军队
	g.AddRouter("dispose", a.dispose, middleware.CheckRole())     //武将配置
	g.AddRouter("conscript", a.conscript, middleware.CheckRole()) //征兵
	g.AddRouter("myOne", a.myOne, middleware.CheckRole())         // 某一个部队详情
	g.AddRouter("assign", a.assign, middleware.CheckRole())       //派遣
}

func (a *ArmyController) myList(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	rspObj := &model.ArmyListRsp{}
	reqObj := &model.ArmyListReq{}
	err := mapstructure.Decode(req.Body.Msg, reqObj)
	if err != nil {
		log.Println("mapstructure decode fail")
		return
	}
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	role, err := req.Conn.GetProperty("role")
	if err != nil {
		rsp.Body.Code = utils.SessionInvalid
		return
	}
	rid := role.(*data.Role).RId
	army, err := logic.RoleArmy.GetRoleArmyByCity(rid, reqObj.CityId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.Armys = army
	rspObj.CityId = reqObj.CityId
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj
}

func (a *ArmyController) dispose(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.DisposeReq{}
	rspObj := &model.DisposeRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = utils.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	//判断一些参数是否合理
	if reqObj.Order < 0 || reqObj.Order > 5 || reqObj.Position < -1 || reqObj.Position > 2 {
		rsp.Body.Code = utils.InvalidParam
		return
	}

	city, ok := logic.RoleCity.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = utils.CityNotExist
		return
	}

	if city.RId != role.RId {
		rsp.Body.Code = utils.CityNotMe
		return
	}

	//校场每升一级一个队伍 比如校场3级 只能有三个队伍
	jc := logic.CityFacilityService.GetFacilityLv(city.CityId, gameConfig.JiaoChang)
	if jc < reqObj.Order || jc <= 0 {
		rsp.Body.Code = utils.ArmyNotEnough
		return
	}
	//查找将领id是否存在
	newG, ok := logic.RoleGeneral.GetByGId(reqObj.GeneralId)
	if ok == false {
		rsp.Body.Code = utils.GeneralNotFound
		return
	}

	if newG.RId != role.RId {
		rsp.Body.Code = utils.GeneralNotMe
		return
	}
	//查询当前位置的军队，没有就创建军队
	army, err := logic.RoleArmy.GetOrCreate(role.RId, reqObj.CityId, reqObj.Order)
	if err != nil {
		rsp.Body.Code = utils.DBError
		return
	}
	//判断军队是不是在城内(是不是出征了)
	if (army.FromX > 0 && army.FromX != city.X) || (army.FromY > 0 && army.FromY != city.Y) {
		rsp.Body.Code = utils.ArmyIsOutside
		return
	}

	//下阵
	if reqObj.Position == -1 { //下
		for pos, g := range army.Gens {
			if g != nil && g.Id == newG.Id {

				//判断武将是否在征兵
				if army.PositionCanModify(pos) == false {
					if army.Cmd == data.ArmyCmdConscript {
						rsp.Body.Code = utils.GeneralBusy
					} else {
						rsp.Body.Code = utils.ArmyBusy
					}
					return
				}
				//下阵
				army.GeneralArray[pos] = 0
				army.SoldierArray[pos] = 0
				army.Gens[pos] = nil
				army.SyncExecute()
				break
			}
		}
		//将领同步
		newG.Order = 0
		newG.CityId = 0
		newG.SyncExecute()
	} else { //上阵
		//征兵中不能上阵
		if army.PositionCanModify(reqObj.Position) == false {
			if army.Cmd == data.ArmyCmdConscript {
				rsp.Body.Code = utils.GeneralBusy
			} else {
				rsp.Body.Code = utils.ArmyBusy
			}
			return
		}
		//已经上阵过了
		if newG.CityId != 0 {
			rsp.Body.Code = utils.GeneralBusy
			return
		}
		//判断是不是上阵过
		if logic.RoleArmy.IsRepeat(role.RId, newG.CfgId) == false {
			rsp.Body.Code = utils.GeneralRepeat
			return
		}

		//判断是否能配前锋 判断统帅厅的等级
		tst := logic.CityFacilityService.GetFacilityLv(city.CityId, gameConfig.TongShuaiTing)
		if reqObj.Position == 2 && (tst < reqObj.Order) {
			rsp.Body.Code = utils.TongShuaiNotEnough
			return
		}

		//判断cost 每个武将有cost 三个加起来不能>cost
		cost := general.General.Cost(newG.CfgId)
		for i, g := range army.Gens {
			if g == nil || i == reqObj.Position {
				continue
			}
			cost += general.General.Cost(g.CfgId)
		}
		//获取city的cost
		if logic.RoleCity.GetCityCost(city.CityId) < cost {
			rsp.Body.Code = utils.CostNotEnough
			return
		}

		oldG := army.Gens[reqObj.Position]
		if oldG != nil {
			//旧的下阵
			oldG.CityId = 0
			oldG.Order = 0
			oldG.SyncExecute()
		}

		//新的上阵
		army.GeneralArray[reqObj.Position] = reqObj.GeneralId
		army.Gens[reqObj.Position] = newG
		army.SoldierArray[reqObj.Position] = 0

		newG.Order = reqObj.Order
		newG.CityId = reqObj.CityId
		newG.SyncExecute()
	}

	army.FromX = city.X
	army.FromY = city.Y
	//army.SyncExecute()
	//队伍
	rspObj.Army = army.ToModel().(model.Army)
	data, _ := json.Marshal(army.GeneralArray)
	army.Generals = string(data)

	data, _ = json.Marshal(army.SoldierArray)
	army.Soldiers = string(data)

	data, _ = json.Marshal(army.ConscriptTimeArray)
	army.ConscriptTimes = string(data)

	data, _ = json.Marshal(army.ConscriptCntArray)
	army.ConscriptCnts = string(data)
	army.SyncExecute()
}

// 征兵
func (a *ArmyController) conscript(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.ConscriptReq{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj := &model.ConscriptRsp{}

	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj
	//征兵army 更新征兵的数量和完成时间 以及状态
	//判断逻辑 征兵能不能进行 资源是否足够 参数是否正常 募兵所等级
	//检查参数
	if len(reqObj.Cnts) != 3 || reqObj.ArmyId <= 0 { //三个角色要征兵
		if reqObj.Cnts[0] < 0 || reqObj.Cnts[1] < 0 || reqObj.Cnts[2] < 0 {
			rsp.Body.Code = utils.InvalidParam
			return
		}
	}
	//登录角色
	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	//查询军队 是否存在
	army, ok := logic.RoleArmy.Get(reqObj.ArmyId)
	if !ok {
		rsp.Body.Code = utils.ArmyNotFound
		return
	}
	//军队是否是归属于此角色
	if role.RId != army.RId {
		rsp.Body.Code = utils.ArmyNotMe
		return
	}
	//判断位置是否可以征兵
	for pos, v := range reqObj.Cnts {
		if v > 0 {
			if army.Gens[pos] == nil {
				rsp.Body.Code = utils.InvalidParam
				return
			}
			//检测武将是否在征兵
			if !army.PositionCanModify(pos) {
				rsp.Body.Code = utils.GeneralBusy
				return
			}
		}
	}
	//募兵所 等级是不是可以征兵
	level := logic.CityFacilityService.GetFacilityLv(army.CityId, gameConfig.MBS)
	if level <= 0 {
		rsp.Body.Code = utils.BuildMBSNotFound
		return
	}
	//是否征兵超限制 根据武将的等级和设施的加成 计算征兵上限 每个武将的等级+设施加成 不能超过上限
	//能征多少兵
	for i, g := range army.Gens {
		if g == nil {
			continue
		}
		//获得这个武将征兵数
		lv := general.GeneralBasic.GetLevel(g.Level)
		if lv == nil {
			rsp.Body.Code = utils.InvalidParam
			return
		}
		//获取城市征兵数量
		add := logic.CityFacilityService.GetSoldier(army.CityId)
		if lv.Soldiers+add < reqObj.Cnts[i]+army.SoldierArray[i] {
			rsp.Body.Code = utils.OutArmyLimit
			return
		}
	}

	//开始征兵 计算消耗资源
	var total int
	for _, v := range reqObj.Cnts {
		total += v
	}
	needRes := gameConfig.NeedRes{
		Decree: 0,
		Gold:   total * gameConfig.Base.ConScript.CostGold,
		Wood:   total * gameConfig.Base.ConScript.CostWood,
		Iron:   total * gameConfig.Base.ConScript.CostIron,
		Grain:  total * gameConfig.Base.ConScript.CostGrain,
		Stone:  total * gameConfig.Base.ConScript.CostStone,
	}
	//判断是否有足够的资源
	code := logic.RoleResService.TryUseNeed(role.RId, needRes)
	if code != utils.OK {
		rsp.Body.Code = code
		return
	}
	//更新部队配置
	for i, _ := range army.SoldierArray {
		var curTime = time.Now().Unix()
		if reqObj.Cnts[i] > 0 {
			army.ConscriptCntArray[i] = reqObj.Cnts[i]
			army.ConscriptTimeArray[i] = int64(reqObj.Cnts[i]*gameConfig.Base.ConScript.CostTime) + curTime - 2
		}
	}
	army.Cmd = data.ArmyCmdConscript
	rspObj.Army = army.ToModel().(model.Army)
	data, _ := json.Marshal(army.GeneralArray)
	army.Generals = string(data)

	data, _ = json.Marshal(army.SoldierArray)
	army.Soldiers = string(data)

	data, _ = json.Marshal(army.ConscriptTimeArray)
	army.ConscriptTimes = string(data)

	data, _ = json.Marshal(army.ConscriptCntArray)
	army.ConscriptCnts = string(data)
	army.SyncExecute()
	rspObj.Army = army.ToModel().(model.Army)
	if res, err := logic.RoleResService.Get(role.RId); err != nil {
		rspObj.RoleRes = res.ToModel().(model.RoleRes)
	}
}

func (a *ArmyController) myOne(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.ArmyOneReq{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj := &model.ArmyOneRsp{}
	rsp.Body.Msg = rspObj
	rsp.Body.Code = utils.OK

	//角色
	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)

	city, ok := logic.RoleCity.Get(reqObj.CityId)
	if !ok {
		rsp.Body.Code = utils.CityNotExist
		return
	}
	if role.RId != city.RId {
		rsp.Body.Code = utils.CityNotMe
		return
	}
	army := logic.RoleArmy.GetArmy(reqObj.CityId, reqObj.Order)
	rspObj.Army = army.ToModel().(model.Army)
}

func (a *ArmyController) assign(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.AssignArmyReq{}
	rspObj := &model.AssignArmyRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = utils.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	//查询军队
	army, ok := logic.RoleArmy.Get(reqObj.ArmyId)
	if ok == false {
		rsp.Body.Code = utils.ArmyNotFound
		return
	}

	if role.RId != army.RId {
		rsp.Body.Code = utils.ArmyNotMe
		return
	}

	if reqObj.Cmd == data.ArmyCmdBack {
		rsp.Body.Code = a.back(army) //回城
	} else if reqObj.Cmd == data.ArmyCmdAttack {
		rsp.Body.Code = a.attack(reqObj, army, role) //占领
	} else if reqObj.Cmd == data.ArmyCmdDefend {
		rsp.Body.Code = a.defend(reqObj, army, role)
	} else if reqObj.Cmd == data.ArmyCmdReclamation {
		rsp.Body.Code = a.reclamation(reqObj, army, role)
	} else if reqObj.Cmd == data.ArmyCmdTransfer {
		rsp.Body.Code = a.transfer(reqObj, army, role)
	}
	rspObj.Army = army.ToModel().(model.Army)
}

func (a *ArmyController) back(army *data.Army) int {
	//从哪里来 回哪里去
	if army.Cmd == data.ArmyCmdAttack ||
		army.Cmd == data.ArmyCmdDefend ||
		army.Cmd == data.ArmyCmdReclamation {
		logic.RoleArmy.ArmyBack(army)
	} else if army.IsIdle() {
		city, ok := logic.RoleCity.Get(army.CityId)
		if ok {
			if city.X != army.FromX || city.Y != army.FromY {
				logic.RoleArmy.ArmyBack(army)
			}
		}

	}
	army.Start = time.Now()
	army.End = time.Now().Add(time.Second * 10)
	logic.RoleArmy.PushAction(army)
	return utils.OK
}

func (a *ArmyController) attack(req *model.AssignArmyReq, army *data.Army, role *data.Role) int {
	if code := a.pre(req, army, role); code != utils.OK {
		return code
	}
	//是否免战 比如刚被占领 不能被攻击（类似新手保护机制）
	if logic.IsWarFree(req.X, req.Y) {
		return utils.BuildWarFree
	}
	//自己的城池 和联盟的城池 都不能攻击
	if logic.IsCanDefend(req.X, req.Y, role.RId) {
		return utils.BuildCanNotAttack
	}
	//计算体力 出征要体力
	power := gameConfig.Base.General.CostPhysicalPower
	for _, v := range army.Gens {
		if v == nil {
			continue
		}
		if v.PhysicalPower < power {
			return utils.PhysicalPowerNotEnough
		}
	}
	//扣除体力
	logic.RoleGeneral.TryUsePhysicalPower(army, power)
	army.ToY = req.Y
	army.ToX = req.X
	army.Cmd = req.Cmd
	army.State = data.ArmyRunning
	now := time.Now()
	army.Start = now
	//实际按照速度来
	army.End = now.Add(time.Second * 10)
	//后台有一个监听程序 一直在看部队是否调动并且到了指定的位置
	//更新缓存
	logic.RoleArmy.PushAction(army)
	return utils.OK
}

func (a *ArmyController) defend(obj *model.AssignArmyReq, army *data.Army, role *data.Role) int {
	return utils.OK
}

func (a *ArmyController) reclamation(obj *model.AssignArmyReq, army *data.Army, role *data.Role) int {
	return utils.OK
}

func (a *ArmyController) transfer(obj *model.AssignArmyReq, army *data.Army, role *data.Role) int {
	return utils.OK
}

func (a *ArmyController) pre(reqObj *model.AssignArmyReq, army *data.Army, role *data.Role) int {
	//判断是否合法
	if reqObj.X < 0 || reqObj.X > globalSet.MapWidth || reqObj.Y < 0 || reqObj.Y > globalSet.MapHeight {
		return utils.InvalidParam
	}
	//是否能出站 是否空闲
	if !army.IsCanOutWar() {
		return utils.ArmyBusy
	}
	if !army.IsIdle() {
		return utils.ArmyBusy
	}
	//判断此土地是否是能攻击的类型 比如山地
	nm, ok := gameConfig.MapRes.ToPositionMap(reqObj.X, reqObj.Y)
	if !ok {
		return utils.InvalidParam
	}
	//山地不能移动到此
	if nm.Type == 0 { //山地
		return utils.InvalidParam
	}
	return utils.OK
}
