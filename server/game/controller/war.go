package controller

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/logic"
	"Three_kingdoms_SLG/server/game/middleware"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"github.com/mitchellh/mapstructure"
	"time"
)

type WarController struct {
}

var DefaultWar = &WarController{}

func (w *WarController) Router(router *net.Router) {
	g := router.Group("war")
	g.Use(middleware.Log())
	g.AddRouter("report", w.report)
	g.AddRouter("read", w.read) //读战报
}

func (w *WarController) report(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//查询战报表 得出数据
	//如果初始化进入游戏，那么要随机初始化三个武将
	rspObj := &model.WarReportRsp{}
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	role, err := req.Conn.GetProperty("role")
	if err != nil {
		rsp.Body.Code = utils.SessionInvalid
		return
	}
	rid := role.(*data.Role).RId
	report, err := logic.RoleWar.GetWarReports(rid)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.List = report
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj
}

func (w *WarController) read(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//查找战报表 得出数据
	reqObj := &model.WarReadReq{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj := &model.WarReadRsp{}
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj
	role, _ := req.Conn.GetProperty("role")
	rid := role.(*data.Role).RId

	rspObj.Id = reqObj.Id

	if reqObj.Id > 0 {
		//更新某一个战报
		//更新进攻方和防守方
		wr := &data.WarReport{
			AttackIsRead:  true,
			DefenseIsRead: true,
			CTime:         time.Now(),
		}
		global.DB.Where("id=? and a_rid=?", reqObj.Id, rid).Create(&wr)

		//global.DB.Where("id=? and d_rid=?", reqObj.Id, rid).Create(&wr)
	} else {
		//更新所有的战报
		wr := &data.WarReport{
			AttackIsRead:  true,
			DefenseIsRead: true,
			CTime:         time.Now(),
		}
		global.DB.Where("id=? and a_rid=?", reqObj.Id, rid).Create(&wr)

		//global.DB.Where("id=? and d_rid=?", reqObj.Id, rid).Create(&wr)
	}
}
