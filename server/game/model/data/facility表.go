package data

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"encoding/json"
	"log"
	"time"
)

type Facility struct {
	Name         string `json:"name"`
	PrivateLevel int8   `json:"level"` //等级，外部读的时候不能直接读，要用GetLevel
	Type         int8   `json:"type"`
	UpTime       int64  `json:"up_time"` //升级的时间戳，0表示该等级已经升级完成了
}

func (f *Facility) GetLevel() int8 {
	if f.UpTime > 0 {
		cur := time.Now().Unix()
		cost := gameConfig.FacilityConf.CostTime(f.Type, f.PrivateLevel+1)
		if cur >= f.UpTime+int64(cost) {
			f.PrivateLevel += 1
			f.UpTime = 0
		}
	}
	return f.PrivateLevel
}

// 更新操作
var CityFacilityDao = &cityFacilityDao{
	cfChan: make(chan *CityFacility),
}

type cityFacilityDao struct {
	cfChan chan *CityFacility
}

func (cf *cityFacilityDao) running() {
	for true {
		select {
		case c := <-cf.cfChan:
			if c.Id > 0 {
				// 使用 GORM 的 `Model` 方法来指定要更新的记录，`Update` 方法指定要更新的字段
				err := global.DB.Model(&c).Select("facilities").Where("id = ?", c.Id).Updates(&c).Error
				if err != nil {
					log.Println("db error", err)
				}
			} else {
				log.Println("update CityFacility fail, because id <= 0")
			}
		}
	}
}

func init() {
	go CityFacilityDao.running()
}
func (f *Facility) CanUp() bool {
	f.GetLevel()
	return f.UpTime == 0
}

type CityFacility struct {
	Id         int    `gorm:"column:id;primaryKey;autoIncrement"`
	RId        int    `gorm:"column:rid"`
	CityId     int    `gorm:"column:cityId"`
	Facilities string `gorm:"column:facilities"` //存的是Facility的字符串
}

func (c *CityFacility) TableName() string {
	return "city_facility"
}

func (c *CityFacility) ChangeFacility() []Facility {
	facilities := make([]Facility, 0)
	err := json.Unmarshal([]byte(c.Facilities), &facilities)
	if err != nil {
		log.Println("unmarshal cityFacility fail")
		return nil
	}
	return facilities
}

func (c *CityFacility) SyncExecute() {
	CityFacilityDao.cfChan <- c
}
