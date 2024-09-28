package controller

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/logic"
	"Three_kingdoms_SLG/server/game/logic/pos"
	"Three_kingdoms_SLG/server/game/middleware"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"log"
	"time"
)

var DefaultRoleController = &RoleController{}

type RoleController struct {
}

func (r *RoleController) Router(router *net.Router) {
	g := router.Group("role")
	g.Use(middleware.Log())
	g.AddRouter("create", r.create)
	g.AddRouter("enterServer", r.enterServer)
	g.AddRouter("myProperty", r.myProperty, middleware.CheckRole())
	g.AddRouter("posTagList", r.posTagList)
	g.AddRouter("upPosition", r.upPosition, middleware.CheckRole()) //实时上报位置，就是在地图上移动 会有位置 要上报

}

func (r *RoleController) enterServer(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	fmt.Println("enter game")
	//进入游戏
	//Session 需要验证是否合法 合法才能取出id
	//根据id取出角色 有就继续玩 没有就创建 然后再查出角色资源，有就返回 没有就初始化
	reqObj := &model.EnterServerReq{}
	rspObj := &model.EnterServerRsp{}
	err := mapstructure.Decode(req.Body.Msg, reqObj)
	//回复的一些信息要和请求对应上
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	if err != nil {
		rsp.Body.Code = utils.InvalidParam
		return
	}
	session := reqObj.Session
	_, claim, err := utils.ParseToken(session)
	if err != nil {
		rsp.Body.Code = utils.SessionInvalid
		return
	}
	uid := claim.Uid

	//进入游戏
	err = logic.RoleService.EnterServer(uid, rspObj, req.Conn)
	if err != nil {
		rspObj.Time = time.Now().UnixNano() / 16
		rsp.Body.Msg = rspObj
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}

	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj
}

func (r *RoleController) myProperty(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//分别根据角色id 去查询 自己有的 东西
	role, err := req.Conn.GetProperty("role") //在Role的EnterServer中已经设置了
	if err != nil {
		rsp.Body.Code = utils.SessionInvalid
		return
	}
	///回复的一些信息要和请求对应上
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	myRole := role.(*data.Role)
	rspObj := &model.MyRolePropertyRsp{}
	//查资源
	rspObj.RoleRes, err = logic.RoleService.GetRoleRes(myRole.RId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	//城池
	rspObj.Citys, err = logic.RoleCity.GetRoleCity(myRole.RId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	//建筑
	rspObj.MRBuilds, err = logic.RoleBuild.GetRoleBuild(myRole.RId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	//军队
	rspObj.Armys, err = logic.RoleArmy.GetRoleArmy(myRole.RId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	//武将
	rspObj.Generals, err = logic.RoleGeneral.GetRoleGeneral(myRole.RId)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	//
	rsp.Body.Msg = rspObj
	rsp.Body.Code = utils.OK
}

func (r *RoleController) posTagList(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	rspObj := &model.PosTagListRsp{}
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	//	去角色属性表查询
	role, err := req.Conn.GetProperty("role")
	if err != nil {
		rsp.Body.Code = utils.SessionInvalid
		return
	}
	rid := role.(*data.Role).RId
	pts, err := logic.RoleAttrService.GetTagList(rid)
	if err != nil {
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.PosTags = pts
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj
}

func (r *RoleController) create(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.CreateRoleReq{}
	rspObj := &model.BuildRoleRsp{}
	err := mapstructure.Decode(req.Body.Msg, reqObj)
	if err != nil {
		log.Println("mapstructure role fail")
		return
	}

	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name
	role := &data.Role{}
	ok := global.DB.Where("uid=?", reqObj.UId).First(&role)
	if ok.Error != nil {
		if ok.Error.Error() == "record not found" {
			// 没有找到记录
			rsp.Body.Code = utils.OK
		} else {
			// 错误发生，可能是数据库连接错误等
			rsp.Body.Code = utils.DBError
			return
		}
	}
	if ok.RowsAffected > 0 {
		// 找到记录
		rsp.Body.Code = utils.RoleAlreadyCreate
		return
	}
	role.UId = reqObj.UId
	role.Sex = reqObj.Sex
	role.NickName = reqObj.NickName
	role.Balance = 0
	role.HeadId = reqObj.HeadId
	role.CreatedAt = time.Now()
	role.LoginTime = time.Now()
	role.LogoutTime = time.Now()
	global.DB.Create(&role)

	rspObj.Role = role.ToModel().(model.Role)
	rsp.Body.Code = utils.OK
	rsp.Body.Msg = rspObj
}

func (rh *RoleController) upPosition(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.UpPositionReq{}
	rspObj := &model.UpPositionRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = utils.OK

	rspObj.X = reqObj.X
	rspObj.Y = reqObj.Y

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	//拿到x y 后上报
	pos.RPMgr.Push(reqObj.X, reqObj.Y, role.RId)
}
