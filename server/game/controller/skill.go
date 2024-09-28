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

var DefaultSkill = &skillController{}

type skillController struct {
}

func (a *skillController) Router(router *net.Router) {
	g := router.Group("skill")
	g.Use(middleware.Log())
	g.AddRouter("list", a.list)
}

func (a *skillController) list(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	rspObj := &model.SkillListRsp{}
	reqObj := &model.SkillListReq{}
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
	skills, err := logic.RoleSkill.GetSkill(rid)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.List = skills
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj
}
