package general

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type gArmsCondition struct {
	Level     int `json:"level"`
	StarLevel int `json:"star_lv"`
}

type gArmsCost struct {
	Gold int `json:"gold"`
}

type gArms struct {
	Id         int            `json:"id"`
	Name       string         `json:"name"`
	Condition  gArmsCondition `json:"condition"`
	ChangeCost gArmsCost      `json:"change_cost"`
	HarmRatio  []int          `json:"harm_ratio"`
}

type Arms struct {
	Title string  `json:"title"`
	Arms  []gArms `json:"arms"`
	AMap  map[int]gArms
}

var GenArms = &Arms{
	AMap: make(map[int]gArms),
}

var generalArmsFile = "/conf/general/general_arms.json"

func (g *Arms) Load() {
	//获取当前文件路径
	currentDir, _ := os.Getwd()
	//配置文件位置
	cf := currentDir + generalArmsFile
	//打包后 程序参数加入配置文件路径
	if len(os.Args) > 1 {
		if path := os.Args[1]; path != "" {
			cf = path + generalArmsFile
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
	for _, v := range g.Arms {
		g.AMap[v.Id] = v
	}
}

func (a *Arms) GetArm(id int) (gArms, error) {
	return a.AMap[id], nil
}

func (a *Arms) GetHarmRatio(attId, defId int) float64 {
	attArm, ok1 := a.AMap[attId]
	_, ok2 := a.AMap[defId]
	if ok1 && ok2 {
		return float64(attArm.HarmRatio[defId-1]) / 100.0
	} else {
		return 1.0
	}
}
