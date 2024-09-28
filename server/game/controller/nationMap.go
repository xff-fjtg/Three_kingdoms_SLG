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

var DefaultNationMap = &nationMapController{}

type nationMapController struct {
}

func (n *nationMapController) Router(router *net.Router) {
	g := router.Group("nationMap")
	g.Use(middleware.Log())
	g.AddRouter("config", n.config)
	//扫描地图
	g.AddRouter("scanBlock", n.scanBlock, middleware.CheckRole())
	g.AddRouter("build", n.build, middleware.CheckRole())   //建设领地
	g.AddRouter("giveUp", n.giveUp, middleware.CheckRole()) //放弃领地 变成空地
}

func (n *nationMapController) config(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	rspObj := &model.ConfigRsp{}
	m := gameConfig.MapBuildConf.Cfg
	rspObj.Confs = make([]model.Conf, len(m))
	for index, v := range m {
		rspObj.Confs[index].Type = v.Type
		rspObj.Confs[index].Name = v.Name
		rspObj.Confs[index].Level = v.Level
		rspObj.Confs[index].Defender = v.Defender
		rspObj.Confs[index].Durable = v.Durable
		rspObj.Confs[index].Grain = v.Grain
		rspObj.Confs[index].Iron = v.Iron
		rspObj.Confs[index].Stone = v.Stone
		rspObj.Confs[index].Wood = v.Wood
	}
	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Code = utils.OK
	rsp.Body.Name = req.Body.Name
	rsp.Body.Msg = rspObj
}

func (n *nationMapController) scanBlock(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//把表中现有的土地性质进行跟变
	reqObj := &model.ScanBlockReq{}
	rspObj := &model.ScanRsp{}
	err := mapstructure.Decode(req.Body.Msg, reqObj)
	if err != nil {
		log.Println("scanBlock Decode error", err)
		return
	}

	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Code = utils.OK
	rsp.Body.Name = req.Body.Name
	//扫描角色建筑 扫描地图
	mrb, err := logic.RoleBuild.ScanBuild(reqObj)
	if err != nil {
		log.Println("ScanBuild  RoleBuild error", err)
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.MRBuilds = mrb
	//角色城池
	mrc, err := logic.RoleCity.ScanBuild(reqObj)
	if err != nil {
		log.Println("ScanBuild RoleCity error", err)
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.MCBuilds = mrc
	role, _ := req.Conn.GetProperty("role")
	rl := role.(*data.Role)
	//扫描玩家军队
	army, err := logic.RoleArmy.ScanBuild(rl.RId, reqObj)
	if err != nil {
		log.Println("ScanBuild  RoleArmy error", err)
		rsp.Body.Code = err.(*common.MyError).Code()
		return
	}
	rspObj.Armys = army
	rsp.Body.Msg = rspObj
}

func (n *nationMapController) build(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &model.BuildReq{}
	rspObj := &model.BuildRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = utils.OK

	x := reqObj.X
	y := reqObj.Y

	rspObj.X = x
	rspObj.Y = y

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	//判断是否是自己的领地 要知道建立什么建筑
	if logic.RoleBuild.BuildIsRId(x, y, role.RId) == false {
		rsp.Body.Code = utils.BuildNotMe
		return
	}

	b, ok := logic.RoleBuild.PositionBuild(x, y)
	if ok == false {
		rsp.Body.Code = utils.BuildNotMe
		return
	}
	//判断是否能建立 如果产出资源不足或者正在建筑 是不能建立的
	if b.IsResBuild() == false || b.IsBusy() {
		rsp.Body.Code = utils.CanNotBuildNew
		return
	}
	//判断建筑是不是到达上限
	cnt := logic.RoleBuild.RoleFortressCnt(role.RId)
	if cnt >= gameConfig.Base.Build.FortressLimit {
		rsp.Body.Code = utils.CanNotBuildNew
		return
	}
	//找到建造要塞 和的所要资源
	cfg, ok := gameConfig.MapBCConf.BuildConfig(reqObj.Type, 1)
	if ok == false {
		rsp.Body.Code = utils.InvalidParam
		return
	}
	//找到资源
	code := logic.RoleResService.TryUseNeed(role.RId, cfg.Need)
	if code != utils.OK {
		rsp.Body.Code = code
		return
	}
	//构建
	b.BuildOrUp(*cfg)
	b.SyncExecute()
}

func (n *nationMapController) giveUp(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//放弃领地 土地要变成系统的
	//放弃有时间 过了时间 才能放弃
	//开一个协程 一直监听  到了时间 才会执行放弃
	reqObj := &model.GiveUpReq{}
	rspObj := &model.GiveUpRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = utils.OK

	x := reqObj.X
	y := reqObj.Y

	rspObj.X = x
	rspObj.Y = y

	r, _ := req.Conn.GetProperty("role")
	role := r.(*data.Role)
	//看看是不是自己的领地
	if logic.RoleBuild.BuildIsRId(x, y, role.RId) == false {
		rsp.Body.Code = utils.BuildNotMe
		return
	}
	//执行放弃
	rsp.Body.Code = logic.RoleBuild.GiveUp(x, y)
}
