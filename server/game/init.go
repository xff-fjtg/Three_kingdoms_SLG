package game

import (
	"Three_kingdoms_SLG/core"
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/game/controller"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/gameConfig/general"
	"Three_kingdoms_SLG/server/game/logic"
)

var Router = &net.Router{}

func Init() {
	global.DB = core.InitGorm()
	//加载基础配置
	gameConfig.Base.Load()

	//加载地图资源配置
	gameConfig.MapBuildConf.Load()

	//加载地图
	gameConfig.MapRes.Load()

	//加载城池设施
	gameConfig.FacilityConf.Load()

	//加载武将军队
	general.General.Load()
	general.GeneralBasic.Load()
	general.GenArms.Load()
	//加载技能
	gameConfig.Skill.Load()
	//加载建筑建设信息
	gameConfig.MapBuildConf.Load()
	//加载地图建筑（升级消耗资源什么的）
	gameConfig.MapBCConf.Load()
	logic.BeforeInit()
	//加载联盟
	logic.CoalitionService.Load()
	//加载所有建筑信息
	logic.RoleBuild.Load()
	//加载所有城池信息
	logic.RoleCity.Load()
	//加载角色属性
	logic.RoleAttrService.Load()
	//加载获取产量
	logic.RoleResService.Load()
	//假设出阵 检测军队是否到达目的地
	logic.RoleArmy.Init()

	InitRouter()
}
func InitRouter() {
	controller.DefaultRoleController.Router(Router)
	controller.DefaultNationMap.Router(Router)
	controller.DefaultGeneral.Router(Router)
	controller.DefaultArmy.Router(Router)
	controller.DefaultWar.Router(Router)
	controller.DefaultSkill.Router(Router)
	controller.InteriorController.Router(Router)
	controller.UnionController.Router(Router)
	controller.CityController.Router(Router)
}
