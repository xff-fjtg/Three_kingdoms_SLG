package controller

import (
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/logic"
	"Three_kingdoms_SLG/server/game/middleware"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"github.com/mitchellh/mapstructure"
	"log"
)

type unionController struct {
}

var UnionController = &unionController{}

func (u *unionController) Router(router *net.Router) {
	g := router.Group("union")
	g.Use(middleware.Log())
	g.AddRouter("list", u.list, middleware.CheckRole())
	g.AddRouter("info", u.info, middleware.CheckRole())
	g.AddRouter("applyList", u.applyList, middleware.CheckRole())
}

func (c *unionController) list(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	rspObj := &model.ListRsp{}
	rsp.Body.Msg = rspObj
	rsp.Body.Code = utils.OK
	//查询数据库，把所有表信息返回
	uns, err := logic.CoalitionService.List()
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.List = uns
	rsp.Body.Msg = rspObj
}

func (c *unionController) info(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.InfoReq{}
	err := mapstructure.Decode(req.Body.Msg, reqObj)
	if err != nil {
		log.Println("info req decode error", err)
		return
	}
	rspObj := &model.InfoRsp{}
	rsp.Body.Msg = rspObj
	rsp.Body.Code = utils.OK
	un, err := logic.CoalitionService.Get(reqObj.Id)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.Info = un
	rspObj.Id = reqObj.Id
	rsp.Body.Msg = rspObj
}

func (u *unionController) applyList(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//根据联盟id 去查询申请列表，rid申请人，你角色表 查询详情即可
	// state 0 正在申请 1 拒绝 2 同意
	//什么人能看到申请列表 只有盟主和副盟主能看到申请列表
	reqObj := &model.ApplyReq{}
	err := mapstructure.Decode(req.Body.Msg, reqObj)
	if err != nil {
		log.Println("applyList req decode error", err)
		return
	}
	rspObj := &model.ApplyRsp{}
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	//查询联盟
	un := logic.CoalitionService.GetCoalition(reqObj.Id)
	if un == nil {
		rsp.Body.Code = utils.DBError
		return
	}
	if un.Chairman != role.RId && un.ViceChairman != role.RId {
		//没有权限
		rspObj.Id = reqObj.Id
		rspObj.Applys = make([]model.ApplyItem, 0)
		return
	}

	ais, err := logic.CoalitionService.GetListApply(reqObj.Id, 0)
	if err != nil {
		rsp.Body.Code = utils.DBError
		return
	}
	rspObj.Id = reqObj.Id
	rspObj.Applys = ais
}
