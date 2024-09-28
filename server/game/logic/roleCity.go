package logic

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/globalSet"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"fmt"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"sync"
	"time"
)

var RoleCity = &roleCityService{
	posRC:  make(map[int]*data.MapRoleCity),
	roleRC: make(map[int][]*data.MapRoleCity),
	dbRB:   make(map[int]*data.MapRoleCity),
}

type roleCityService struct {
	dbRB  map[int]*data.MapRoleCity
	mutex sync.RWMutex
	//位置 为key  建筑id 为value
	posRC map[int]*data.MapRoleCity
	//key 为角色id value为map  key 为位置 value 为建筑
	roleRC map[int][]*data.MapRoleCity
}

func (r *roleCityService) Load() {
	//查询所有角色建筑

	dbRB := make(map[int]*data.MapRoleCity)
	var roleCities []*data.MapRoleCity

	// 查询所有 MapRoleCity 数据
	result := global.DB.Find(&roleCities).Error
	if result != nil {
		log.Println("加载玩家城池出错:", result)
		return
	}

	// 将查询结果转换为 map
	for _, city := range roleCities {
		dbRB[city.RId] = city
	}
	for _, v := range dbRB {
		posId := globalSet.ToPosition(v.X, v.Y)
		r.posRC[posId] = v
		_, ok := r.roleRC[v.RId]
		if !ok {
			r.roleRC[v.RId] = make([]*data.MapRoleCity, 0)
		}
		//创建之后再加
		r.roleRC[v.RId] = append(r.roleRC[v.RId], v)
	}
}

func (r *roleCityService) InitCity(rid int, name string, conn net.WSConn, tx *gorm.DB) error {
	roleCity := &data.MapRoleCity{}
	count := tx.Where("rid = ?", rid).First(&roleCity).RowsAffected
	//找角色
	if count == 0 {
		for {
			//没有城池 初始化 条件系统城市5格内不能有玩家城池
			roleCity.X = rand.Intn(globalSet.MapWidth)
			roleCity.Y = rand.Intn(globalSet.MapHeight)
			//判断是否符合创建条件 五格子之内不能有别的玩家城池
			if r.ISCanBuild(roleCity.X, roleCity.Y) {
				//TODO
				//建的肯定是主城
				roleCity.RId = rid
				roleCity.CreatedAt = time.Now()
				roleCity.Name = name
				roleCity.IsMain = 1
				roleCity.CurDurable = gameConfig.Base.City.Durable
				err := tx.Create(&roleCity).Error
				if err != nil {
					log.Println("插入玩家城市出错", err)
					tx.Rollback()
					return common.New(utils.DBError, "插入玩家城市出错")
				}
				//新创建城池加入缓存
				posId := globalSet.ToPosition(roleCity.X, roleCity.Y)
				r.posRC[posId] = roleCity
				_, ok := r.roleRC[rid]
				if !ok {
					r.roleRC[rid] = make([]*data.MapRoleCity, 0)
				} else {
					r.roleRC[rid] = append(r.roleRC[rid], roleCity)
				}
				r.dbRB[roleCity.CityId] = roleCity
				//TODO 生成主城后初始化设施信息 查询有没有 有就初始化 没有就算了
				if err := CityFacilityService.TryCreate(roleCity.CityId, rid, tx); err != nil {
					log.Println("insert 城池facility fail", err)
					tx.Rollback()
					return common.New(err.(*common.MyError).Code(), err.Error())
				}
				break
			}
		}

	}
	r.dbRB[roleCity.CityId] = roleCity
	return nil
}

func (r *roleCityService) ISCanBuild(x int, y int) bool {
	confs := gameConfig.MapRes.Confs
	pIndex := globalSet.ToPosition(x, y)
	_, ok := confs[pIndex]
	if !ok {
		return false
	}

	//城池 1范围内 不能超过边界
	if x+1 >= globalSet.MapHeight || y+1 >= globalSet.MapHeight || y-1 < 0 || x-1 < 0 {
		return false
	}
	sysBuild := gameConfig.MapRes.SysBuild
	//系统城池的5格内 不能创建玩家城池
	for _, v := range sysBuild {
		if v.Type == gameConfig.MapBuildSysCity {
			if x >= v.X-5 &&
				x <= v.X+5 &&
				y >= v.Y-5 &&
				y <= v.Y+5 {
				return false
			}
		}
	}

	//玩家城池的5格内 也不能创建城池
	for i := x - 5; i <= x+5; i++ {
		for j := y - 5; j <= y+5; j++ {
			posId := globalSet.ToPosition(i, j)
			_, ok := r.posRC[posId]
			if ok {
				return false
			}
		}
	}
	return true
}

func (r *roleCityService) GetRoleCity(rid int) ([]model.MapRoleCity, error) {
	cities := make([]data.MapRoleCity, 0)
	err := global.DB.Where("rid = ?", rid).Find(&cities).Error
	modelCities := make([]model.MapRoleCity, 0)
	if err != nil {
		log.Println("search role cities fail", err)
		return modelCities, err
	}
	for _, v := range cities {
		modelCities = append(modelCities, v.ToModel().(model.MapRoleCity))
	}
	return modelCities, nil
}

func (r *roleCityService) ScanBuild(req *model.ScanBlockReq) ([]model.MapRoleCity, error) {
	x := req.X
	y := req.Y
	length := req.Length
	maxX := utils.MaxInt(globalSet.MapWidth, x+length-1)
	maxY := utils.MaxInt(globalSet.MapHeight, y+length-1)

	rb := make([]model.MapRoleCity, 0)
	if x < 0 || x >= maxX || y < 0 || y >= maxY {
		return rb, nil
	}

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {
			posId := globalSet.ToPosition(i, j)
			v, ok := r.posRC[posId]
			if ok {
				rb = append(rb, v.ToModel().(model.MapRoleCity))
			}
		}
	}
	fmt.Println("success!!!!!!!!!!!")
	return rb, nil
}

func (r *roleCityService) Get(id int) (*data.MapRoleCity, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	city, ok := r.dbRB[id]
	if ok {
		return city, ok
	}
	return nil, ok
}

func (r *roleCityService) GetMainCity(rid int) (*data.MapRoleCity, bool) {
	rcs, ok := r.roleRC[rid]
	if !ok {
		return nil, false
	}
	for _, v := range rcs {
		if v.IsMain == 1 {
			return v, true
		}
	}
	return nil, false
}

func (r *roleCityService) GetCityCost(cid int) int8 {
	return CityFacilityService.GetCost(cid) + gameConfig.Base.City.Cost
}

func (r *roleCityService) PositionCity(x int, y int) (*data.MapRoleCity, bool) {
	pos := globalSet.ToPosition(x, y)
	rb, ok := r.posRC[pos]
	if ok {
		return rb, true
	}
	return nil, false
}
