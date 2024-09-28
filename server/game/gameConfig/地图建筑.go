package gameConfig

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var MapBuildConf = &mapBuildConf{
	cfgMap: make(map[int8][]cfg),
}

type mapBuildConf struct {
	Title  string `json:"title"`
	Cfg    []cfg  `json:"cfg"`
	cfgMap map[int8][]cfg
}

// 读取资源
const mapBuildConfFile = "/conf/game/map_build.json"

func (m *mapBuildConf) Load() {
	//获取当前文件路径
	currentDir, _ := os.Getwd()
	//配置文件位置
	cf := currentDir + mapBuildConfFile
	//打包后 程序参数加入配置文件路径
	if len(os.Args) > 1 {
		if path := os.Args[1]; path != "" {
			cf = path + mapBuildConfFile
		}
	}
	data, err := ioutil.ReadFile(cf)
	if err != nil {
		log.Println("地图配置资源读取失败")
		panic(err)
	}
	err = json.Unmarshal(data, m)
	if err != nil {
		log.Println("地图配置资源格式定义失败")
		panic(err)
	}
	for _, v := range m.Cfg {
		_, ok := m.cfgMap[v.Type]
		if !ok {
			m.cfgMap[v.Type] = make([]cfg, 0)
		}
		m.cfgMap[v.Type] = append(m.cfgMap[v.Type], v)
	}
}

func (m *mapBuildConf) BuildConfig(buildType int8, level int8) *cfg {
	cfgs := m.cfgMap[buildType]
	for _, v := range cfgs {
		if v.Level == level {
			return &v
		}
	}
	return nil
}

type cfg struct { //占领了 可以加多少
	Type     int8   `json:"type"`
	Name     string `json:"name"`
	Level    int8   `json:"level"`
	Grain    int    `json:"grain"`
	Wood     int    `json:"wood"`
	Iron     int    `json:"iron"`
	Stone    int    `json:"stone"`
	Durable  int    `json:"durable"`
	Defender int    `json:"defender"`
}
