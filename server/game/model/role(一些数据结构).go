package model

import "Three_kingdoms_SLG/server/game/gameConfig"

type EnterServerReq struct {
	Session string `json:"session"`
}
type MyRolePropertyReq struct {
}

// 相应
type EnterServerRsp struct {
	Role    Role    `json:"role"`
	RoleRes RoleRes `json:"role_res"`
	Time    int64   `json:"time"`
	Token   string  `json:"token"`
}

type Role struct { //角色信息
	RId      int    `json:"rid"`
	UId      int    `json:"uid"`
	NickName string `json:"nickName"`
	Sex      int8   `json:"sex"`
	Balance  int    `json:"balance"`
	HeadId   int16  `json:"headId"`
	Profile  string `json:"profile"`
}

type RoleRes struct { //角色资源
	Wood          int `json:"wood"`
	Iron          int `json:"iron"`
	Stone         int `json:"stone"`
	Grain         int `json:"grain"`
	Gold          int `json:"gold"`
	Decree        int `json:"decree"`     //令牌
	WoodYield     int `json:"wood_yield"` //资源增长量
	IronYield     int `json:"iron_yield"`
	StoneYield    int `json:"stone_yield"`
	GrainYield    int `json:"grain_yield"`
	GoldYield     int `json:"gold_yield"`     //征收
	DepotCapacity int `json:"depot_capacity"` //仓库容量
}
type PosTagListRsp struct {
	PosTags []PosTag `json:"pos_tags"`
}
type PosTag struct { //位置信息
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Name string `json:"name"`
}
type MyGeneralRsp struct { //武将回复
	Generals []General `json:"generals"`
}
type GSkill struct { //武将技能
	Id    int `json:"id"`
	Lv    int `json:"lv"`
	CfgId int `json:"cfgId"`
}
type MyRolePropertyRsp struct {
	RoleRes  RoleRes        `json:"role_res"`
	MRBuilds []MapRoleBuild `json:"mr_builds"` //角色建筑，包含被占领的基础建筑
	Generals []General      `json:"generals"`  //角色武将
	Citys    []MapRoleCity  `json:"citys"`     //角色城池
	Armys    []Army         `json:"armys"`     //角色军队
}
type MapRoleBuild struct {
	RId        int    `json:"rid"`
	RNick      string `json:"RNick"` //角色昵称
	Name       string `json:"name"`
	UnionId    int    `json:"union_id"`   //联盟id
	UnionName  string `json:"union_name"` //联盟名字
	ParentId   int    `json:"parent_id"`  //上级id
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Type       int8   `json:"type"`
	Level      int8   `json:"level"`
	OPLevel    int8   `json:"op_level"`
	CurDurable int    `json:"cur_durable"`
	MaxDurable int    `json:"max_durable"`
	Defender   int    `json:"defender"`
	OccupyTime int64  `json:"occupy_time"`
	EndTime    int64  `json:"end_time"`    //建造完的时间
	GiveUpTime int64  `json:"giveUp_time"` //领地到了这个时间会被放弃
}

func (b *MapRoleBuild) IsRoleFortress() bool {
	return b.Type == gameConfig.MapBuildFortress
}

type General struct { //武将
	Id            int       `json:"id"`
	CfgId         int       `json:"cfgId"` //用的是那个武将
	PhysicalPower int       `json:"physical_power"`
	Order         int8      `json:"order"`
	Level         int8      `json:"level"`
	Exp           int       `json:"exp"`
	CityId        int       `json:"cityId"`
	CurArms       int       `json:"curArms"`
	HasPrPoint    int       `json:"hasPrPoint"`
	UsePrPoint    int       `json:"usePrPoint"`
	AttackDis     int       `json:"attack_distance"`
	ForceAdded    int       `json:"force_added"`
	StrategyAdded int       `json:"strategy_added"`
	DefenseAdded  int       `json:"defense_added"`
	SpeedAdded    int       `json:"speed_added"`
	DestroyAdded  int       `json:"destroy_added"`
	StarLv        int8      `json:"star_lv"`
	Star          int8      `json:"star"`
	ParentId      int       `json:"parentId"`
	Skills        []*GSkill `json:"skills"`
	State         int8      `json:"state"`
}

func (g *General) ToArray() []int {
	r := make([]int, 0)
	r = append(r, g.Id)
	r = append(r, g.CfgId)
	r = append(r, g.PhysicalPower)
	r = append(r, int(g.Order))
	r = append(r, int(g.Level))
	r = append(r, g.Exp)
	r = append(r, g.CityId)
	r = append(r, g.CurArms)
	r = append(r, g.HasPrPoint)
	r = append(r, g.UsePrPoint)
	r = append(r, g.AttackDis)
	r = append(r, g.ForceAdded)
	r = append(r, g.StrategyAdded)
	r = append(r, g.DefenseAdded)
	r = append(r, g.SpeedAdded)
	r = append(r, g.DestroyAdded)
	r = append(r, int(g.StarLv))
	r = append(r, int(g.Star))
	return r
}

type MapRoleCity struct {
	CityId     int    `json:"cityId"`
	RId        int    `json:"rid"`
	Name       string `json:"name"`
	UnionId    int    `json:"union_id"`   //联盟id
	UnionName  string `json:"union_name"` //联盟名字
	ParentId   int    `json:"parent_id"`  //上级id
	X          int    `json:"x"`
	Y          int    `json:"y"`
	IsMain     bool   `json:"is_main"`
	Level      int8   `json:"level"`
	CurDurable int    `json:"cur_durable"`
	MaxDurable int    `json:"max_durable"`
	OccupyTime int64  `json:"occupy_time"`
}
type Army struct {
	Id       int     `json:"id"`
	CityId   int     `json:"cityId"`
	UnionId  int     `json:"union_id"` //联盟id
	Order    int8    `json:"order"`    //第几队，1-5队
	Generals []int   `json:"generals"`
	Soldiers []int   `json:"soldiers"`
	ConTimes []int64 `json:"con_times"`
	ConCnts  []int   `json:"con_cnts"`
	Cmd      int8    `json:"cmd"`
	State    int8    `json:"state"` //状态:0:running,1:stop
	FromX    int     `json:"from_x"`
	FromY    int     `json:"from_y"`
	ToX      int     `json:"to_x"`
	ToY      int     `json:"to_y"`
	Start    int64   `json:"start"` //出征开始时间
	End      int64   `json:"end"`   //出征结束时间
}
type CreateRoleReq struct {
	UId      int    `json:"uid"`
	NickName string `json:"nickName"`
	Sex      int8   `json:"sex"`
	SId      int    `json:"sid"`
	HeadId   int16  `json:"headId"`
}
type BuildRoleRsp struct {
	Role Role `json:"role"`
}

// 上报位置
type UpPositionReq struct {
	X int `json:"x"`
	Y int `json:"y"`
}
type UpPositionRsp struct {
	X int `json:"x"`
	Y int `json:"y"`
}
