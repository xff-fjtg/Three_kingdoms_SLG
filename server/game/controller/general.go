package controller

import (
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/logic"
	"Three_kingdoms_SLG/server/game/middleware"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"github.com/mitchellh/mapstructure"
	"log"
)

var DefaultGeneral = &GeneralController{}

type GeneralController struct {
}

func (n *GeneralController) Router(router *net.Router) {
	g := router.Group("general")
	g.Use(middleware.Log())
	g.AddRouter("myGenerals", n.myGenerals)   //初始化武将
	g.AddRouter("drawGeneral", n.drawGeneral) //抽卡
}

func (n *GeneralController) myGenerals(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//查询自己的已有的武将
	//如果初始化进入游戏，那么要随机初始化三个武将
	rspObj := &model.MyGeneralRsp{}
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	role, err := req.Conn.GetProperty("role")
	if err != nil {
		rsp.Body.Code = utils.SessionInvalid
		return
	}
	rid := role.(*data.Role).RId
	gs, err := logic.RoleGeneral.GetRoleGeneral(rid)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.Generals = gs
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj
}

func (n *GeneralController) drawGeneral(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//1. 计算抽卡花费的金钱
	//2. 判断金钱是否足够
	//3. 抽卡的次数 + 已有的武将 卡池是否足够
	//4. 随机生成武将即可（之前有实现）
	//5. 金币的扣除
	reqObj := &model.DrawGeneralReq{}
	err := mapstructure.Decode(req.Body.Msg, reqObj)
	if err != nil {
		log.Println("drawGeneral req decode error", err)
		return
	}
	rspObj := &model.DrawGeneralRsp{}
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj
	role, _ := req.Conn.GetProperty("role")
	rid := role.(*data.Role).RId
	//抽卡金币的计算
	cost := gameConfig.Base.General.DrawGeneralCost * reqObj.DrawTimes
	if !logic.RoleResService.IsEnoughGold(rid, cost) {
		rsp.Body.Code = utils.GoldNotEnough //钱不够
		return
	}
	limit := gameConfig.Base.General.Limit
	//拿武将
	gs, err := logic.RoleGeneral.GetRoleGeneral(rid)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	if len(gs)+reqObj.DrawTimes > limit {
		rsp.Body.Code = utils.OutGeneralLimit
		return
	}
	//抽卡
	mgs := logic.RoleGeneral.Draw(rid, reqObj.DrawTimes)
	logic.RoleResService.CostGold(rid, cost)
	rspObj.Generals = mgs
}
