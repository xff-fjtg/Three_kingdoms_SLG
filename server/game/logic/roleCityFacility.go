package logic

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"encoding/json"
	"gorm.io/gorm"
	"log"
	"sync"
	"time"
)

var CityFacilityService = &cityFacilityService{}

type cityFacilityService struct {
	mutex sync.RWMutex
}

// 用于尝试根据给定的城市 ID (cid) 和角色 ID (rid) 创建一个 CityFacility 记录和相关的 Facility 记录。
func (c *cityFacilityService) TryCreate(cid, rid int, tx *gorm.DB) error {

	// 初始化一个 CityFacility 结构体指针
	cf := &data.CityFacility{}

	// 在数据库中查找给定 cid 的 CityFacility 记录，并获取受影响的行数。如果找到记录，count 将大于 0。
	count := tx.Where("cityId = ?", cid).First(&cf)

	// 如果有数据的话，直接返回。就是说，数据库中已经存在具有给定 cid 的 CityFacility 记录，函数直接返回 nil，表示不需要进行任何操作。
	if count.RowsAffected > 0 {
		return nil
		// 否则，使用 common.New 函数创建一个新的错误，错误类型为 utils.DBError，错误信息为 "cityFacility already exist"，表示尝试创建一个已经存在的 CityFacility 记录。
	}

	// 如果没有找到记录，程序将继续执行，为 cf 结构体设置 CityId 和 RId 属性，分别对应于函数参数中的 cid 和 rid。
	cf.CityId = cid
	cf.RId = rid
	// 然后，程序初始化了一个新的切片 list，其长度与 gameConfig.FacilityConf.List 相同。
	list := gameConfig.FacilityConf.List
	// 并创建一个相应长度的 Facility 切片
	facs := make([]data.Facility, len(list))
	// 接着，程序遍历 list 切片，为每个元素创建一个对应的 Facility 结构体。
	for index, v := range list {
		// 创建城池设施
		// 先创建城池设施
		// 初始化一个 Facility 结构体，设置 Type 和 Name 属性，对应于 gameConfig.FacilityConf.List 中的元素。PrivateLevel 和 UpTime 属性被设置为默认值 0。
		fac := data.Facility{
			Type:         v.Type,
			Name:         v.Name,
			PrivateLevel: 0,
			UpTime:       0,
		}
		// 将新创建的 Facility 结构体添加到 facs 切片中。
		facs[index] = fac
	}
	// 将初始化后的 pieces 切片 JSON 序列化后存储到 cf.Facilities 字段中，用于记录角色拥有的所有棋子的初始信息。
	dataJson, _ := json.Marshal(facs)
	cf.Facilities = string(dataJson)

	// 最后使用 global.DB.Create(&cf) 尝试在数据库中创建新的 CityFacility 记录。如果创建成功，函数返回 nil。如果创建失败，函数检查 err.Error 是否为 nil，使用 common.New 函数创建一个新的错误，错误类型为 utils.DBError，错误信息为 "insert cityFacility fail"，表示在尝试创建 CityFacility 记录时失败。
	err := tx.Create(&cf)
	if err.Error != nil {
		tx.Rollback()
		return common.New(utils.DBError, "insert cityFacility fail")
	}
	log.Println("TryCreate cityFacility success")
	return nil
}

func (c *cityFacilityService) GetByRId(rid int) ([]*data.CityFacility, error) {
	cf := make([]*data.CityFacility, 0)
	err := global.DB.Where("rid = ?", rid).Find(&cf)
	if err.Error != nil {
		log.Println(err)
		return cf, common.New(utils.DBError, "数据库错误")
	}
	return cf, nil
}

func (c *cityFacilityService) GetYield(rid int) data.Yield {
	//把表中的设施找到（facility表） 然后根据不同的类型 去配置中找到，匹配到增加产量的设施
	//木头什么的 计算出来
	//设施的等级不同 产量也不同
	cfs, err := c.GetByRId(rid)
	var y data.Yield
	if err == nil {
		for _, cf := range cfs {
			for _, fa := range cf.ChangeFacility() { //转换一下类型
				if fa.GetLevel() > 0 {
					//计算等级 不同等级资源产出不同
					values := gameConfig.FacilityConf.GetValues(fa.Type, fa.GetLevel())
					additions := gameConfig.FacilityConf.GetAdditions(fa.Type)
					for i, aType := range additions { //不同additions有不同作用
						if aType == gameConfig.TypeWood {
							y.Wood += values[i]
						} else if aType == gameConfig.TypeGrain {
							y.Grain += values[i]
						} else if aType == gameConfig.TypeIron {
							y.Iron += values[i]
						} else if aType == gameConfig.TypeStone {
							y.Stone += values[i]
						} else if aType == gameConfig.TypeTax {
							y.Gold += values[i]
						}
					}
				}
			}
		}
	}
	log.Println("GetYield yes")
	return y
}

func (c *cityFacilityService) GetFacility(rid, cid int) ([]data.Facility, error) {
	cf := &data.CityFacility{}
	err := global.DB.Where("rid = ? and cityId = ?", rid, cid).Find(&cf)
	if err.Error != nil {
		log.Println(err)
		return nil, common.New(utils.DBError, "数据库错误")
	}
	return cf.ChangeFacility(), nil
}
func (c *cityFacilityService) GetFacility2(rid, cid int) (*data.CityFacility, error) {
	cf := &data.CityFacility{}
	err := global.DB.Where("rid = ? and cityId = ?", rid, cid).Find(&cf)
	if err.Error != nil {
		log.Println(err)
		return nil, common.New(utils.DBError, "数据库错误")
	}
	return cf, nil
}
func (c *cityFacilityService) UpFacility(rid, cid int, fType int8) (*data.Facility, int) {
	c.mutex.RLock()
	f, ok := c.GetFacility2(rid, cid)
	c.mutex.RUnlock()

	if ok != nil {
		return nil, utils.CityNotExist
	} else {
		facilities := make([]*data.Facility, 0)
		var out *data.Facility
		json.Unmarshal([]byte(f.Facilities), &facilities)
		for _, fac := range facilities {
			if fac.Type == fType {
				//找到设施 判断能否升级
				maxLevel := gameConfig.FacilityConf.MaxLevel(fType)
				if fac.CanUp() == false {
					//正在升级中了
					return nil, utils.UpError
				} else if fac.GetLevel() >= maxLevel {
					return nil, utils.UpError
				} else {
					//判断要多少资源 在判断用户有多少资源
					need, ok := gameConfig.FacilityConf.Need(fType, fac.GetLevel()+1)
					if ok == false {
						return nil, utils.UpError
					}

					code := RoleResService.TryUseNeed(rid, *need)
					if code == utils.OK {
						fac.UpTime = time.Now().Unix()
						out = fac
						if t, err := json.Marshal(facilities); err == nil {
							f.Facilities = string(t)
							f.SyncExecute()
							return out, utils.OK
						} else {
							return nil, utils.UpError
						}
					} else {
						return nil, code
					}
				}
			}
		}

		return nil, utils.UpError
	}
}
func (c *cityFacilityService) GetByCid(cid int) (*data.CityFacility, error) {
	cf := &data.CityFacility{}
	err := global.DB.Where("cityId = ? ", cid).Find(&cf)
	if err.Error != nil {
		log.Println(err)
		return nil, common.New(utils.DBError, "数据库错误")
	}
	return cf, nil
}
func (c *cityFacilityService) GetFacilityLv(cid int, shi int8) int8 {
	cf, _ := c.GetByCid(cid)
	if cf == nil {
		return 0
	}
	facs := cf.ChangeFacility()
	for _, v := range facs {
		if v.Type == shi {
			return v.GetLevel()
		}
	}
	return 0
}

func (c *cityFacilityService) GetCost(cid int) int8 {
	cf, _ := c.GetByCid(cid)
	facilitys := cf.ChangeFacility()
	var cost int
	for _, fa := range facilitys { //转换一下类型
		if fa.GetLevel() > 0 {
			//计算等级 不同等级资源产出不同
			values := gameConfig.FacilityConf.GetValues(fa.Type, fa.GetLevel())
			additions := gameConfig.FacilityConf.GetAdditions(fa.Type)
			for i, aType := range additions { //不同additions有不同作用
				if aType == gameConfig.TypeCost {
					cost += values[i]
				}
			}
		}
	}
	return int8(cost)
}

func (c *cityFacilityService) GetDepotCapacity(rid int) int {
	cfs, err := c.GetByRId(rid)
	limit := 0
	if err == nil {
		for _, cf := range cfs {
			for _, f := range cf.ChangeFacility() {
				if f.GetLevel() > 0 {
					values := gameConfig.FacilityConf.GetValues(f.Type, f.GetLevel())
					additions := gameConfig.FacilityConf.GetAdditions(f.Type)
					for i, aType := range additions {
						if aType == gameConfig.TypeWarehouseLimit {
							limit += values[i]
						}
					}
				}
			}
		}
	}
	return limit
}

func (c *cityFacilityService) GetSoldier(cid int) int {
	cf, ok := c.GetByCid(cid)
	limit := 0
	if ok == nil {
		for _, f := range cf.ChangeFacility() {
			if f.GetLevel() > 0 {
				values := gameConfig.FacilityConf.GetValues(f.Type, f.GetLevel())
				additions := gameConfig.FacilityConf.GetAdditions(f.Type)
				for i, aType := range additions {
					if aType == gameConfig.TypeSoldierLimit {
						limit += values[i]
					}
				}
			}
		}
	}
	return limit
}

func (c *cityFacilityService) GetAdditions(cid int, additionType ...int8) []int {
	cf, _ := c.GetByCid(cid)
	ret := make([]int, len(additionType))
	if cf == nil {
		return ret
	} else {
		for i, at := range additionType {
			limit := 0
			for _, f := range cf.ChangeFacility() {
				if f.GetLevel() > 0 {
					values := gameConfig.FacilityConf.GetValues(f.Type, f.GetLevel())
					additions := gameConfig.FacilityConf.GetAdditions(f.Type)
					for i, aType := range additions {
						if aType == at {
							limit += values[i]
						}
					}
				}
			}
			ret[i] = limit
		}
	}
	return ret
}
