package logic

import "Three_kingdoms_SLG/server/game/model/data"

func BeforeInit() {
	//接口赋值
	data.GetYield = RoleService.GetYield
	data.GetUnion = RoleAttrService.GetUnion
	data.GetParentId = RoleAttrService.GetParentId
	data.MapResTypeLevel = RoleBuild.MapResTypeLevel
	data.GetMainMembers = CoalitionService.GetMainMembers
}

//使 data.GetYield 成为一个指向 RoleService.GetYield 的函数引用。
