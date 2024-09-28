package controller

import (
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/game/logic"
	"Three_kingdoms_SLG/server/game/middleware"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"github.com/mitchellh/mapstructure"
)

var CityController = &cityController{}

type cityController struct {
}

func (c *cityController) Router(router *net.Router) {
	g := router.Group("city")
	g.Use(middleware.Log())
	g.AddRouter("facilities", c.facilities, middleware.CheckRole())
	g.AddRouter("upFacility", c.upFacility, middleware.CheckRole())

}

func (c *cityController) facilities(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.FacilitiesReq{}
	rspObj := &model.FacilitiesRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.CityId = reqObj.CityId
	rsp.Body.Code = utils.OK
	//得到角色
	r, _ := req.Conn.GetProperty("role")
	////查询城池
	city, ok := logic.RoleCity.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = utils.CityNotExist
		return
	}

	role := r.(*data.Role)
	if city.RId != role.RId {
		rsp.Body.Code = utils.CityNotMe
		return
	}
	//查询城池设施
	f, err := logic.CityFacilityService.GetFacility(role.RId, reqObj.CityId)
	if err != nil {
		rsp.Body.Code = utils.CityNotExist
		return
	}

	rspObj.Facilities = make([]model.Facility, len(f))
	for i, v := range f {
		rspObj.Facilities[i].Name = v.Name
		rspObj.Facilities[i].Level = v.GetLevel()
		rspObj.Facilities[i].Type = v.Type
		rspObj.Facilities[i].UpTime = v.UpTime
	}
}

func (c *cityController) upFacility(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//1.根据城池id 查询城池和设施 保证城池存在
	//2.需要更新upTime 升级完 uptime=0
	//3.要判断符合条件 一些资源什么的 符合了才能升级 升级完成要更新数据库
	//4.消耗的资源也要更新回数据库
	//5.查询的资源要放回前端
	reqObj := &model.UpFacilityReq{}
	rspObj := &model.UpFacilityRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.CityId = reqObj.CityId
	rsp.Body.Code = utils.OK

	r, _ := req.Conn.GetProperty("role")
	city, ok := logic.RoleCity.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = utils.CityNotExist
		return
	}
	//判断角色是不是自己
	role := r.(*data.Role)
	if city.RId != role.RId {
		rsp.Body.Code = utils.CityNotMe
		return
	}
	//判断设施
	facs, _ := logic.CityFacilityService.GetFacility(role.RId, reqObj.CityId)
	if facs == nil {
		rsp.Body.Code = utils.CityNotExist
		return
	}

	out, errCode := logic.CityFacilityService.UpFacility(role.RId, reqObj.CityId, reqObj.FType)
	rsp.Body.Code = errCode
	if errCode == utils.OK {
		rspObj.Facility.Level = out.GetLevel()
		rspObj.Facility.Type = out.Type
		rspObj.Facility.Name = out.Name
		rspObj.Facility.UpTime = out.UpTime

		if roleRes, ok := logic.RoleResService.Get(role.RId); ok == nil {
			rspObj.RoleRes = roleRes.ToModel().(model.RoleRes)
		}
	}
	rspObj.CityId = reqObj.CityId
}
