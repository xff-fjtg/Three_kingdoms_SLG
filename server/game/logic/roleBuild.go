package logic

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/globalSet"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"log"
	"sync"
	"time"
)

type roleBuildService struct {
	giveUpMutex sync.RWMutex
	mutex       sync.RWMutex
	//位置 为key  建筑id 为value
	posRB map[int]*data.MapRoleBuild
	//key 为角色id value为map  key 为位置 value 为建筑
	roleRB map[int][]*data.MapRoleBuild
	//放弃 时间 建筑
	giveUpRB map[int64]map[int]*data.MapRoleBuild
}

var RoleBuild = &roleBuildService{
	giveUpRB: make(map[int64]map[int]*data.MapRoleBuild),
	posRB:    make(map[int]*data.MapRoleBuild),
	roleRB:   make(map[int][]*data.MapRoleBuild),
}

func (r *roleBuildService) Load() {
	//加载系统建筑和玩家建筑
	//先需要判断数据库是否保存了系统建筑，没有要进行保存
	result := global.DB.Where("type = ? or type = ?", gameConfig.MapBuildSysCity, gameConfig.MapBuildFortress)
	if int64(len(gameConfig.MapRes.SysBuild)) != result.RowsAffected {
		//	对不上，需要将系统建筑存入数据库
		//		先删除 后插入
		err := global.DB.Where("type = ? or type = ?", gameConfig.MapBuildSysCity, gameConfig.MapBuildSysFortress).Delete(new(data.MapRoleBuild)).Error
		if err != nil {
			log.Println("delete MapRoleBuild fail")
			return
		}
		for _, v := range gameConfig.MapRes.SysBuild {
			build := data.MapRoleBuild{
				RId:        0,
				Type:       v.Type,
				Level:      v.Level,
				X:          v.X,
				Y:          v.Y,
				EndTime:    time.Now(),
				OccupyTime: time.Now(),
			}
			build.Init()
			global.DB.Create(&build)
		}
	}
	//查询所有角色建筑
	//dbRB := make(map[int]*data.MapRoleBuild)
	//err := global.DB.Find(&dbRB).Error
	// 先用 slice 查询结果
	var dbRBList []data.MapRoleBuild

	// 执行查询，将结果填充到 slice 中
	err := global.DB.Find(&dbRBList).Error
	if err != nil {
		log.Println("查询错误: ", err)
		return
	}

	// 创建 map，并将 slice 的结果转换成 map 形式
	dbRB := make(map[int]*data.MapRoleBuild)
	for _, v := range dbRBList {
		dbRB[v.Id] = &v
	}
	// 现在 dbRB 里就有从数据库查询到的数据了

	for _, v := range dbRB {
		v.Init()
		posId := globalSet.ToPosition(v.X, v.Y)
		r.posRB[posId] = v
		_, ok := r.roleRB[v.RId]
		if !ok {
			r.roleRB[v.RId] = make([]*data.MapRoleBuild, 0)
		}
		r.roleRB[v.RId] = append(r.roleRB[v.RId], v)
	}
	//放弃的也要加载
	for _, v := range dbRB {
		v.Init()
		if v.GiveUpTime > 0 {
			_, ok := r.giveUpRB[v.GiveUpTime]
			if !ok {
				r.giveUpRB[v.GiveUpTime] = make(map[int]*data.MapRoleBuild)
			}
			r.giveUpRB[v.GiveUpTime][v.Id] = v
		}
	}
	go r.checkGiveUp()
}
func (r *roleBuildService) GetRoleBuild(rid int) ([]model.MapRoleBuild, error) {
	roleBuild := make([]data.MapRoleBuild, 0)
	err := global.DB.Where("rid = ?", rid).Find(&roleBuild).Error
	if err != nil {
		log.Println("search roleBuild fail")
		return nil, common.New(utils.DBError, "search roleBuild fail")
	}
	//modelBuild := make([]model.MapRoleBuild, len(roleBuild))
	modelBuild := make([]model.MapRoleBuild, 0)
	//如果不是0，在append的时候就变成len= len(roleBuild)+len(roleBuild)了
	for _, v := range roleBuild {
		modelBuild = append(modelBuild, v.ToModel().(model.MapRoleBuild))
	}
	return modelBuild, nil
}

func (r *roleBuildService) ScanBuild(req *model.ScanBlockReq) ([]model.MapRoleBuild, error) {
	x := req.X
	y := req.Y
	length := req.Length
	mrbs := make([]model.MapRoleBuild, 0)
	if x < 0 || x >= globalSet.MapWidth || y < 0 || y >= globalSet.MapHeight {
		return mrbs, nil
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	maxX := utils.MinInt(globalSet.MapWidth, x+length-1)
	maxY := utils.MinInt(globalSet.MapHeight, y+length-1)
	//是一个范围，要在x-length 到 x+length 之间(y也一样)
	for i := x - length; i <= maxX; i++ {
		for j := y - length; j <= maxY; j++ {
			posId := globalSet.ToPosition(x, y)
			mrb, ok := r.posRB[posId]
			if ok {
				mrbs = append(mrbs, mrb.ToModel().(model.MapRoleBuild))
			}
		}
	}
	return mrbs, nil
}

func (r *roleBuildService) GetYield(rid int) data.Yield {
	var y data.Yield
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	builds, ok := r.roleRB[rid]
	if ok {
		for _, b := range builds {
			y.Iron += b.Iron
			y.Wood += b.Wood
			y.Grain += b.Grain
			y.Stone += b.Grain
		}
	}
	return y
}

func (r *roleBuildService) PositionBuild(x int, y int) (*data.MapRoleBuild, bool) {
	pos := globalSet.ToPosition(x, y)
	rb, ok := r.posRB[pos]
	if ok {
		return rb, true
	}
	return nil, false
}

func (r *roleBuildService) BuildCnt(rid int) int {
	rbs, ok := r.roleRB[rid]
	if ok {
		return len(rbs)
	}
	return 0
}

func (r *roleBuildService) RemoveFromRole(build *data.MapRoleBuild) {
	rb, ok := r.roleRB[build.RId]
	if ok {
		for i, v := range rb {
			if v.Id == build.Id {
				r.roleRB[build.RId] = append(rb[:i], rb[i+1:]...)
				break
			}
		}
	}
	r.giveUpMutex.Lock()
	delete(r.giveUpRB, build.GiveUpTime)
	r.giveUpMutex.Unlock()
	//重制成系统的
	build.Reset()
	build.SyncExecute()
}
func (r *roleBuildService) MapResTypeLevel(x int, y int) (bool, int8, int8) {
	posId := globalSet.ToPosition(x, y)
	rb, ok := gameConfig.MapRes.Confs[posId]
	if ok {
		return true, rb.Type, rb.Level
	}
	return false, 0, 0
}

func (r *roleBuildService) AddBuild(rid int, x int, y int) (*data.MapRoleBuild, bool) {
	posId := globalSet.ToPosition(x, y)
	rb, ok := r.posRB[posId]
	if ok {
		if r.roleRB[rid] == nil {
			r.roleRB[rid] = make([]*data.MapRoleBuild, 0)
		}
		r.roleRB[rid] = append(r.roleRB[rid], rb)

		return rb, true
	} else {
		//数据库插入
		if b, ok := gameConfig.MapRes.PositionBuild(x, y); ok { //找到对应的土地
			if cfg := gameConfig.MapBuildConf.BuildConfig(b.Type, b.Level); cfg != nil {
				rb := &data.MapRoleBuild{
					RId: rid, X: x, Y: y,
					Type: b.Type, Level: b.Level, OPLevel: b.Level,
					Name: cfg.Name, CurDurable: cfg.Durable,
					MaxDurable: cfg.Durable,
					OccupyTime: time.Now(),
					EndTime:    time.Now().Add(time.Duration(cfg.Durable) * time.Second),
				}
				rb.Init()

				if result := global.DB.Create(&rb); result.Error == nil {
					r.posRB[posId] = rb
					if _, ok := r.roleRB[rid]; ok == false {
						r.roleRB[rid] = make([]*data.MapRoleBuild, 0)
					}
					r.roleRB[rid] = append(r.roleRB[rid], rb)
					return rb, true
				}

			}
		}
	}
	return nil, false
}

func (r *roleBuildService) RoleFortressCnt(rid int) int {
	bs, err := r.GetRoleBuild(rid)
	cnt := 0
	if err != nil {
		return 0
	} else {
		for _, b := range bs {
			//有一个玩家要塞 就加上1
			if b.IsRoleFortress() {
				cnt += 1
			}
		}
	}
	return cnt
}

func (r *roleBuildService) BuildIsRId(x int, y int, rid int) bool {
	build, ok := r.PositionBuild(x, y)
	if ok {
		if build.RId == rid {
			return true
		}
	}
	return false
}

func (r *roleBuildService) GiveUp(x int, y int) int {
	b, ok := r.PositionBuild(x, y)
	if ok == false {
		return utils.CannotGiveUp
	}
	//打仗不能放弃
	if b.IsWarFree() {
		return utils.BuildWarFree
	}

	if b.GiveUpTime > 0 {
		return utils.BuildGiveUpAlready
	}
	//放弃时间=当前时间+系统配置的时间
	b.GiveUpTime = time.Now().Unix() + gameConfig.Base.Build.GiveUpTime
	b.SyncExecute()

	_, ok = r.giveUpRB[b.GiveUpTime]
	if ok == false {
		r.giveUpRB[b.GiveUpTime] = make(map[int]*data.MapRoleBuild)
	}
	r.giveUpRB[b.GiveUpTime][b.Id] = b

	return utils.OK
}

func (r *roleBuildService) checkGiveUp() {
	for {
		time.Sleep(time.Second * 2)
		var ret []int
		var builds []*data.MapRoleBuild
		r.giveUpMutex.RLock()
		cur := time.Now().Unix()
		for i := cur - 10; i <= cur; i++ {
			gs, ok := r.giveUpRB[i]

			if ok {
				for _, g := range gs {
					//要放弃的build
					builds = append(builds, g)
					//放弃土地后 现在要有部队 要返回
					ret = append(ret, globalSet.ToPosition(g.X, g.Y))
				}
			}
		}
		r.giveUpMutex.RUnlock()

		for _, build := range builds {
			r.RemoveFromRole(build)
		}
		for _, posId := range ret {
			RoleArmy.GiveUpPosId(posId)
		}
	}
}
