package gameConfig

import (
	"Three_kingdoms_SLG/server/game/globalSet"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type NationalMap struct {
	MId   int  `gorm:"column:mid"`
	X     int  `gorm:"column:x"`
	Y     int  `gorm:"column:y"`
	Type  int8 `gorm:"column:type"`
	Level int8 `gorm:"column:level"`
}

type mapRes struct { //地图资源
	Confs    map[int]NationalMap //每个格子的定义
	SysBuild map[int]NationalMap //系统建筑
}

type mapData struct {
	Width  int     `json:"w"`
	Height int     `json:"h"`
	List   [][]int `json:"list"`
}

const (
	MapBuildSysFortress = 50 //系统要塞
	MapBuildSysCity     = 51 //系统城市
	MapBuildFortress    = 56 //玩家要塞
)

var MapRes = &mapRes{
	Confs:    make(map[int]NationalMap),
	SysBuild: make(map[int]NationalMap),
}

const mapFile = "/conf/game/map.json"

func (m *mapRes) Load() {
	//获取当前文件路径
	currentDir, _ := os.Getwd()
	//配置文件位置
	cf := currentDir + mapFile
	//打包后 程序参数加入配置文件路径
	if len(os.Args) > 1 {
		if path := os.Args[1]; path != "" {
			cf = path + mapFile
		}
	}
	data, err := ioutil.ReadFile(cf)
	if err != nil {
		log.Println("地图读取失败")
		panic(err)
	}
	mapData := &mapData{}
	err = json.Unmarshal(data, mapData)
	if err != nil {
		log.Println("地图格式定义失败")
		panic(err)
	}
	globalSet.MapHeight = mapData.Height
	globalSet.MapWidth = mapData.Width
	for index, v := range mapData.List {
		t := int8(v[0])
		l := int8(v[1]) //土地等级
		nm := NationalMap{
			X:     index % globalSet.MapWidth,
			Y:     index / globalSet.MapHeight,
			Type:  t,
			Level: l,
			MId:   index,
		}
		m.Confs[index] = nm
		if t == MapBuildSysCity || t == MapBuildSysFortress {
			m.SysBuild[index] = nm
		}
	}

}
func (m *mapRes) ToPositionMap(x, y int) (NationalMap, bool) {
	position := globalSet.ToPosition(x, y)
	nm, ok := MapRes.Confs[position]
	if ok {
		return nm, true
	}
	return nm, false
}
func (m *mapRes) PositionBuild(x, y int) (NationalMap, bool) {
	position := globalSet.ToPosition(x, y)
	nm, ok := MapRes.Confs[position]
	if ok {
		return nm, true
	}
	return nm, false
}

func (m *mapRes) IsCanBuild(x int, y int) bool {
	posIndex := globalSet.ToPosition(x, y)
	c, ok := m.Confs[posIndex]
	if ok {
		//type==0是一个山地 不能被占领
		if c.Type == 0 {
			return false
		} else {
			return true
		}
	} else {
		return false
	}
}
