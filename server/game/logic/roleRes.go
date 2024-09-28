package logic

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"errors"
	"gorm.io/gorm"
	"log"
	"time"
)

var RoleResService = &roleResService{
	rolesRes: make(map[int]*data.RoleRes),
}

type roleResService struct {
	rolesRes map[int]*data.RoleRes
}

func (r *roleResService) Get(rid int) (*data.RoleRes, error) {
	ra := &data.RoleRes{}
	result := global.DB.Where("rid = ?", rid).First(&ra) // 或者使用 Take(&ra)
	// 检查是否发生错误
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 如果没有找到记录
			log.Println("查询角色资源出错", result.Error)
			return ra, common.New(utils.DBError, "查询角色资源出错")
		}
		// 其他数据库错误
		log.Println(result.Error)
		return ra, common.New(utils.DBError, "数据库错误")
	}
	return ra, nil
}

func (r *roleResService) IsEnoughGold(rid int, cost int) bool {
	rr, err := r.Get(rid)
	if err != nil {
		log.Println("IsEnoughGold 查询角色资源出错", err)
		return false
	}
	return rr.Gold >= cost
}

func (r *roleResService) CostGold(rid int, cost int) {
	rr, _ := r.Get(rid)
	if rr.Gold >= cost {
		rr.Gold -= cost
		rr.SyncExecute()
	}
}

func (r *roleResService) TryUseNeed(rid int, need gameConfig.NeedRes) int {
	rr, err := r.Get(rid)

	if err == nil {
		if need.Decree <= rr.Decree && need.Grain <= rr.Grain &&
			need.Stone <= rr.Stone && need.Wood <= rr.Wood &&
			need.Iron <= rr.Iron && need.Gold <= rr.Gold {
			rr.Decree -= need.Decree
			rr.Iron -= need.Iron
			rr.Wood -= need.Wood
			rr.Stone -= need.Stone
			rr.Grain -= need.Grain
			rr.Gold -= need.Gold

			rr.SyncExecute()
			return utils.OK
		} else {
			if need.Decree > rr.Decree {
				return utils.DecreeNotEnough
			} else {
				return utils.ResNotEnough
			}
		}
	} else {
		return utils.RoleNotExist
	}
}
func (r *roleResService) Load() {
	rr := make([]*data.RoleRes, 0)
	err := global.DB.Find(&rr)
	if err != nil {
		log.Println(" load role_res table error")
	}

	for _, v := range rr {
		r.rolesRes[v.RId] = v
	}

	go r.produce()
}

//获取产量

func (r *roleResService) produce() {
	index := 1
	for true {
		//一直去获取产量 隔一段时间就刷新一次
		t := gameConfig.Base.Role.RecoveryTime
		time.Sleep(time.Duration(t) * time.Second)

		for _, v := range r.rolesRes {
			//加判断是因为爆仓了，资源不无故减少
			capacity := GetDepotCapacity(v.RId)
			//增加产量
			yield := data.GetYield(v.RId)
			if v.Wood < capacity {
				v.Wood += utils.MinInt(yield.Wood/6, capacity)
			}

			if v.Iron < capacity {
				v.Iron += utils.MinInt(yield.Iron/6, capacity)
			}

			if v.Stone < capacity {
				v.Stone += utils.MinInt(yield.Stone/6, capacity)
			}

			if v.Grain < capacity {
				v.Grain += utils.MinInt(yield.Grain/6, capacity)
			}

			if v.Gold < capacity {
				v.Grain += utils.MinInt(yield.Grain/6, capacity)
			}
			//恢复令牌
			if index%6 == 0 {
				if v.Decree < gameConfig.Base.Role.DecreeLimit {
					v.Decree += 1
				}
			}
			v.SyncExecute()
		}
		index++
	}
}

func GetDepotCapacity(rid int) int {
	return CityFacilityService.GetDepotCapacity(rid) + gameConfig.Base.Role.DepotCapacity
}
