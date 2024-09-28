package general

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type generalBasic struct {
	Title  string   `json:"title"`
	Levels []gLevel `json:"levels"`
}
type gLevel struct {
	Level    int8 `json:"level"`
	Exp      int  `json:"exp"`
	Soldiers int  `json:"soldiers"` //带兵的数量
}

var GeneralBasic = &generalBasic{}
var generalBasicFile = "/conf/general/general_basic.json"

func (g *generalBasic) Load() {
	//获取当前文件路径
	currentDir, _ := os.Getwd()
	//配置文件位置
	cf := currentDir + generalBasicFile
	//打包后 程序参数加入配置文件路径
	if len(os.Args) > 1 {
		if path := os.Args[1]; path != "" {
			cf = path + generalBasicFile
		}
	}
	data, err := ioutil.ReadFile(cf)
	if err != nil {
		log.Println("武将基本配置读取失败")
		panic(err)
	}
	err = json.Unmarshal(data, g)
	if err != nil {
		log.Println("武将基本配置格式定义失败")
		panic(err)
	}
}

func (g *generalBasic) GetLevel(level int8) *gLevel {
	for _, v := range g.Levels {
		if v.Level == level {
			return &v
		}
	}
	return nil
}

func (g *generalBasic) ExpToLevel(exp int) (int8, int) {
	var level int8 = 0
	//总共要的经验
	limitExp := g.Levels[len(g.Levels)-1].Exp
	for _, v := range g.Levels {
		if exp >= v.Exp && v.Level > level {
			level = v.Level
		}
	}

	if limitExp < exp {
		return level, limitExp
	} else {
		return level, exp
	}
}
