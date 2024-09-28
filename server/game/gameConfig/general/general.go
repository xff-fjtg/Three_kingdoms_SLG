package general

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
)

type general struct { //武将表
	Title            string          `json:"title"`
	GArr             []generalDetail `json:"list"`
	GMap             map[int]generalDetail
	totalProbability int
}
type generalDetail struct {
	Name         string `json:"name"`
	CfgId        int    `json:"cfgId"`
	Force        int    `json:"force"`    //武力
	Strategy     int    `json:"strategy"` //策略
	Defense      int    `json:"defense"`  //防御
	Speed        int    `json:"speed"`    //速度
	Destroy      int    `json:"destroy"`  //破坏力
	ForceGrow    int    `json:"force_grow"`
	StrategyGrow int    `json:"strategy_grow"`
	DefenseGrow  int    `json:"defense_grow"`
	SpeedGrow    int    `json:"speed_grow"`
	DestroyGrow  int    `json:"destroy_grow"`
	Cost         int8   `json:"cost"`
	Probability  int    `json:"probability"`
	Star         int8   `json:"star"`
	Arms         []int  `json:"arms"`
	Camp         int8   `json:"camp"`
}

var General = &general{}
var generalFile = "/conf/general/general.json"

func (g *general) Load() {
	g.GArr = make([]generalDetail, 0)
	g.GMap = make(map[int]generalDetail)
	//获取当前文件路径
	currentDir, _ := os.Getwd()
	//配置文件位置
	configFile := currentDir + generalFile
	//打包后 程序参数加入配置文件路径
	if len(os.Args) > 1 {
		if path := os.Args[1]; path != "" {
			configFile = path + generalFile
		}
	}
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("facility读取失败")
		panic(err)
	}
	err = json.Unmarshal(data, g)
	if err != nil {
		log.Println("facility json 转换失败")
		panic(err)
	}
	for _, v := range g.GArr {
		g.GMap[v.CfgId] = v
		g.totalProbability += v.Probability //抽取概率总和
	}
}

// 随机武将
func (g *general) RandomGeneral() int {
	//随机抽取
	rate := rand.Intn(g.totalProbability)
	var cur = 0
	//随机抽取武将
	for _, v := range g.GArr {
		if rate >= cur && rate < cur+v.Probability {
			return v.CfgId
		}
		cur += v.Probability
	}
	return 0
}

func (g *general) Cost(cfgId int) int8 {
	c, ok := g.GMap[cfgId]
	if ok {
		return c.Cost
	} else {
		return 0
	}
}
