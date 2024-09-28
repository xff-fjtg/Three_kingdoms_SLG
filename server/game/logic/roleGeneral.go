package logic

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/gameConfig/general"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"
)

// 武将出征需要消耗体力，定期需要补充体力
type roleGeneralService struct {
	mutex     sync.RWMutex
	genByRole map[int][]*data.General
	genByGId  map[int]*data.General
}

var RoleGeneral = &roleGeneralService{
	genByRole: make(map[int][]*data.General),
	genByGId:  make(map[int]*data.General),
}

func (r *roleGeneralService) GetRoleGeneral(rid int) ([]model.General, error) {
	roleGeneral := make([]*data.General, 0)
	err := global.DB.Where("rid = ?", rid).Find(&roleGeneral).Error
	if err != nil {
		log.Println("search roleGeneral fail")
		return nil, common.New(utils.DBError, "search roleBuild fail")
	}
	if len(roleGeneral) <= 0 {
		//没有，随机三个
		count := 0
		for {
			if count >= 3 {
				break
			}
			cfgId := general.General.RandomGeneral()
			gen, err := r.NewGeneral(cfgId, rid, 1)
			if err != nil {
				log.Println(err)
				continue
			}

			roleGeneral = append(roleGeneral, gen)
			count++
		}
	}

	modelGeneral := make([]model.General, 0)
	for _, v := range roleGeneral {
		modelGeneral = append(modelGeneral, v.ToModel().(model.General))
	}
	return modelGeneral, nil
}

const (
	GeneralNormal      = 0 //正常
	GeneralComposeStar = 1 //星级合成
	GeneralConvert     = 2 //转换
)

func (r *roleGeneralService) NewGeneral(cfgId int, rid int, level int) (*data.General, error) {
	cfg := general.General.GMap[cfgId]
	sa := make([]*model.GSkill, 3) //三个技能槽
	ss, _ := json.Marshal(sa)      //函数转换为 JSON 格式
	gen := &data.General{
		PhysicalPower: gameConfig.Base.General.PhysicalPowerLimit,
		RId:           rid,
		CfgId:         cfg.CfgId,
		Order:         0,
		CityId:        0,
		Level:         int8(level),
		CreatedAt:     time.Now(),
		CurArms:       cfg.Arms[0],
		HasPrPoint:    0,
		UsePrPoint:    0,
		AttackDis:     0,
		ForceAdded:    0,
		StrategyAdded: 0,
		DefenseAdded:  0,
		SpeedAdded:    0,
		DestroyAdded:  0,
		Star:          cfg.Star,
		StarLv:        0,
		ParentId:      0,
		SkillsArray:   sa,
		Skills:        string(ss),
		State:         GeneralNormal,
	}
	//钩子
	data, _ := json.Marshal(gen.SkillsArray)
	gen.Skills = string(data)
	result := global.DB.Create(&gen)
	if result.Error != nil {
		log.Println("create roleGeneral fail")
		return nil, result.Error
	}
	return gen, nil
}

func (g *roleGeneralService) Draw(rid int, nums int) []model.General {
	mrs := make([]*data.General, 0)
	for i := 0; i < nums; i++ {
		cfgId := general.General.RandomGeneral()
		gen, _ := g.NewGeneral(cfgId, rid, 1) //这里保存在数据库中了
		mrs = append(mrs, gen)
	}
	modelMrs := make([]model.General, 0)
	for _, v := range mrs {
		modelMrs = append(modelMrs, v.ToModel().(model.General))
	}
	return modelMrs
}

func (g *roleGeneralService) GetByGId(id int) (*data.General, bool) {
	gen := &data.General{}
	result := global.DB.Where("id = ? and state = ? ", id, data.GeneralNormal).Find(&gen)
	if result.Error != nil {
		log.Println(result.Error)
		return nil, false
	}

	return gen, true

}

func (g *roleGeneralService) Load() {

	var generals []*data.General

	// 查询符合条件的 generals 并填充到 generals 切片中
	result := global.DB.Table("generals").
		Where("state = ?", data.GeneralNormal).
		Find(&generals)

	if result.Error != nil {
		log.Println("Error:", result.Error)
		return
	}

	// 加锁，保证多线程安全
	RoleGeneral.mutex.Lock()
	defer RoleGeneral.mutex.Unlock()

	// 清空现有数据（根据实际需求决定是否清空）
	RoleGeneral.genByGId = make(map[int]*data.General)
	RoleGeneral.genByRole = make(map[int][]*data.General)

	// 填充 genByGId，并更新 genByRole
	for _, general := range generals {
		// 更新 genByGId
		RoleGeneral.genByGId[general.Id] = general

		// 初始化 genByRole，如果 genByRole[general.RId] 还没有值
		if _, ok := RoleGeneral.genByRole[general.RId]; !ok {
			RoleGeneral.genByRole[general.RId] = make([]*data.General, 0)
		}
		// 将当前 general 加入 genByRole[general.RId]
		RoleGeneral.genByRole[general.RId] = append(RoleGeneral.genByRole[general.RId], general)
	}
	go g.updatePhysicalPower()
}

func (g *roleGeneralService) updatePhysicalPower() {
	limit := gameConfig.Base.General.PhysicalPowerLimit
	recoverCnt := gameConfig.Base.General.RecoveryPhysicalPower
	for true {
		time.Sleep(1 * time.Hour)
		g.mutex.RLock()
		for _, gen := range g.genByGId {
			// 恢复体力
			if gen.PhysicalPower < limit {
				gen.PhysicalPower = utils.MinInt(limit, gen.PhysicalPower+recoverCnt)
				gen.SyncExecute()
			}
		}
		g.mutex.RUnlock()
	}
}

func (g *roleGeneralService) TryUsePhysicalPower(army *data.Army, power int) {
	for _, v := range army.Gens {
		if v == nil {
			continue
		}
		v.PhysicalPower -= power
		v.SyncExecute()
	}
}

func (g *roleGeneralService) GetDestroy(army *data.Army) int {
	//所有武将的破坏力
	destroy := 0
	for _, gen := range army.Gens {
		if gen == nil {
			continue
		}
		destroy += gen.GetDestroy()
	}
	return destroy
}

// 获取npc武将
func (gen *roleGeneralService) GetNPCGenerals(cnt int, star int8, level int8) ([]data.General, bool) {
	//获取系统的武将
	gs, ok := gen.GetByRId(0)
	if ok == false {
		return make([]data.General, 0), false
	} else {
		target := make([]data.General, 0)
		for _, g := range gs {
			if g.Level == level && g.Star == star {
				target = append(target, *g)
			}
		}

		if len(target) < cnt {
			return make([]data.General, 0), false
		} else {
			m := make(map[int]int)
			for true {
				r := rand.Intn(len(target))
				m[r] = r
				if len(m) == cnt {
					break
				}
			}

			rgs := make([]data.General, 0)
			for _, v := range m {
				t := target[v]
				rgs = append(rgs, t)
			}
			return rgs, true
		}
	}
}

func (g *roleGeneralService) GetByRId(rid int) ([]*data.General, bool) {
	mrs := make([]*data.General, 0)
	result := global.DB.Where("rid = ? ", rid).Find(&mrs)
	if result.Error != nil {
		log.Println("武将查询出错", result.Error)
		return nil, false
	}
	if len(mrs) <= 0 {
		//随机三个武将 做为初始武将
		var count = 0
		for {
			if count >= 3 {
				break
			}
			cfgId := general.General.RandomGeneral()
			if cfgId != 0 {
				gen, err := g.NewGeneral(cfgId, rid, 1)
				if err != nil {
					log.Println("生成武将出错", err)
					continue
				}
				mrs = append(mrs, gen)
				count++
			}
		}

	}
	return mrs, true
}
