package controller

import (
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/logic"
	"Three_kingdoms_SLG/server/game/middleware"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"log"
	"time"
)

// 在role_attribute里面
var InteriorController = &interiorController{}

type interiorController struct {
}

func (i *interiorController) Router(router *net.Router) {
	g := router.Group("interior")
	g.Use(middleware.Log())
	g.AddRouter("openCollect", i.openCollect, middleware.CheckRole())
	g.AddRouter("collect", i.collect, middleware.CheckRole())
	g.AddRouter("transform", i.transform, middleware.CheckRole())
}
func (i *interiorController) openCollect(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	rspObj := &model.OpenCollectionrsp{}

	rsp.Body.Msg = rspObj
	rsp.Body.Code = utils.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	roleAttr, err := logic.RoleAttrService.Get(role.RId)
	if err != nil {
		//征收次数
		rsp.Body.Code = utils.DBError
		return
	}

	rspObj.Limit = gameConfig.Base.Role.CollectTimesLimit
	rspObj.CurTimes = roleAttr.CollectTimes
	interval := gameConfig.Base.Role.CollectInterval
	if roleAttr.LastCollectTime.IsZero() {
		rspObj.NextTime = 0
	} else {
		if roleAttr.CollectTimes >= rspObj.Limit {
			//今天已经完成 下次征收就是第二天
			//第二天从0点开始
			y, m, d := roleAttr.LastCollectTime.Add(24 * time.Hour).Date()
			//东八区time.FixedZone（CST）
			nextTime := time.Date(y, m, d, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
			rspObj.NextTime = nextTime.UnixNano() / 1e6
		} else {
			nextTime := roleAttr.LastCollectTime.Add(time.Duration(interval) * time.Second)
			rspObj.NextTime = nextTime.UnixNano() / 1e6
		}
	}
}

func (i *interiorController) collect(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//查询角色资源 获得当前金币
	//查询角色属性 获取征收信息
	//查询获取当前产量 征收的金币是多少
	rspObj := &model.CollectionRsp{}

	rsp.Body.Msg = rspObj
	rsp.Body.Code = utils.OK
	//资源
	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	roleRes, err := logic.RoleResService.Get(role.RId)
	if err != nil {
		rsp.Body.Code = utils.DBError
		return
	}
	//属性
	roleAttr, err := logic.RoleAttrService.Get(role.RId)
	if err != nil {
		rsp.Body.Code = utils.DBError
		return
	}

	curTime := time.Now()
	lastTime := roleAttr.LastCollectTime
	if curTime.YearDay() != lastTime.YearDay() || curTime.Year() != lastTime.Year() {
		roleAttr.CollectTimes = 0
		roleAttr.LastCollectTime = time.Time{}
	}

	timeLimit := gameConfig.Base.Role.CollectTimesLimit
	//是否超过征收次数上限
	if roleAttr.CollectTimes >= timeLimit {
		rsp.Body.Code = utils.OutCollectTimesLimit
		return
	}

	//cd内不能操作
	need := lastTime.Add(time.Duration(gameConfig.Base.Role.CollectTimesLimit) * time.Second)
	if curTime.Before(need) {
		rsp.Body.Code = utils.InCdCanNotOperate
		return
	}
	//获取产量
	gold := logic.RoleService.GetYield(roleRes.RId).Gold
	rspObj.Gold = gold
	roleRes.Gold += gold
	//SyncExecute channel更新
	roleRes.SyncExecute()
	//计算征收
	roleAttr.LastCollectTime = curTime
	roleAttr.CollectTimes += 1
	roleAttr.SyncExecute()

	interval := gameConfig.Base.Role.CollectInterval
	if roleAttr.CollectTimes >= timeLimit {
		y, m, d := roleAttr.LastCollectTime.Add(24 * time.Hour).Date()
		nextTime := time.Date(y, m, d, 0, 0, 0, 0, time.FixedZone("IST", 3600))
		rspObj.NextTime = nextTime.UnixNano() / 1e6
	} else {
		nextTime := roleAttr.LastCollectTime.Add(time.Duration(interval) * time.Second)
		rspObj.NextTime = nextTime.UnixNano() / 1e6
	}

	rspObj.CurTimes = roleAttr.CollectTimes
	rspObj.Limit = timeLimit
}

func (i *interiorController) transform(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//查询资源
	//查询集市是否符合要求
	//from to from减去 to增加
	reqObj := &model.TransformReq{}
	rspObj := &model.TransformRsp{}

	err := mapstructure.Decode(req.Body.Msg, reqObj)
	if err != nil {
		log.Println("transform decode err", err)
		return
	}
	rsp.Body.Msg = rspObj
	rsp.Body.Code = utils.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	roleRes, err := logic.RoleResService.Get(role.RId)
	if err != nil {
		rsp.Body.Code = utils.DBError
		return
	}

	//做交易的时候，主城做交易
	main, _ := logic.RoleCity.GetMainCity(role.RId)

	//判断等级
	lv := logic.CityFacilityService.GetFacilityLv(main.CityId, gameConfig.JiShi)
	if lv <= 0 {
		rsp.Body.Code = utils.NotHasJiShi
		return
	}
	//四个东西 从哪里to到哪里
	fmt.Println("yes!!!!", reqObj.To, reqObj.From)
	lens := 4
	fmt.Println("yes!!!!", reqObj.To, reqObj.From, lens)
	ret := make([]int, lens)
	fmt.Println("yes!!!!", reqObj.To, reqObj.From)
	for i := 0; i < lens; i++ {
		//ret[i] = reqObj.To[i] - reqObj.From[i]
		//from什么东西>0 就要减去
		fmt.Println("ret")
		if reqObj.From[i] > 0 {
			ret[i] = -reqObj.From[i]
		}

		//to什么东西>0 就要加
		if reqObj.To[i] > 0 {
			ret[i] = reqObj.To[i]
		}
	}
	//检查资源是否够 不管怎样 ret要么>0 要么<0
	//检查每个资源加上变化量后是否小于 0： - 如果 Wood 加上 ret[0] 后小于 0，则返回 utils.InvalidParam。 - 类似地检查 Iron、Stone 和 Grain。
	if roleRes.Wood+ret[0] < 0 {
		rsp.Body.Code = utils.InvalidParam
		return
	}

	if roleRes.Iron+ret[1] < 0 {
		rsp.Body.Code = utils.InvalidParam
		return
	}

	if roleRes.Stone+ret[2] < 0 {
		rsp.Body.Code = utils.InvalidParam
		return
	}

	if roleRes.Grain+ret[3] < 0 {
		rsp.Body.Code = utils.InvalidParam
		return
	}
	//都够 就执行
	roleRes.Wood += ret[0]
	roleRes.Iron += ret[1]
	roleRes.Stone += ret[2]
	roleRes.Grain += ret[3]
	roleRes.SyncExecute()
}
