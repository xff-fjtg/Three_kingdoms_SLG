package gameConfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type basic struct {
	ConScript conscript `json:"conscript"`
	General   general   `json:"general"`
	Role      role      `json:"role"`
	City      city      `json:"city"`
	Npc       npc       `json:"npc"`
	Union     union     `json:"union"`
	Build     build     `json:"build"`
}

var Base = &basic{}

const basicFile = "/conf/game/basic.json"

// 读取资源
func (b *basic) Load() {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	configPath := currentDir + basicFile
	if !fileExist(configPath) {
		panic("basic.json is not exist")
	}
	Len := len(os.Args)
	if Len > 1 {
		dir := os.Args[1]
		if dir != "" {
			configPath = dir + basicFile
		}
	}
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic("set basic.json fail")
	}
	err = json.Unmarshal(data, b)
	if err != nil {
		panic("Unmarshal basic.json fail")
	}
	fmt.Println("Load file success")
}

func (b *basic) GetNPC(l int8) (npcLevel, bool) {
	levels := b.Npc.Levels
	for index, v := range levels {
		if index+1 == int(l) {
			return v, true
		}
	}
	return npcLevel{}, false
}

func fileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil || os.IsExist(err)
}

type conscript struct { //征兵
	Des       string `json:"des"`
	CostWood  int    `json:"cost_wood"`
	CostIron  int    `json:"cost_iron"`
	CostStone int    `json:"cost_stone"`
	CostGrain int    `json:"cost_grain"`
	CostGold  int    `json:"cost_gold"`
	CostTime  int    `json:"cost_time"` //每征一个兵需要花费时间
}

type general struct {
	Des                   string `json:"des"`
	PhysicalPowerLimit    int    `json:"physical_power_limit"`    //体力上限
	CostPhysicalPower     int    `json:"cost_physical_power"`     //消耗体力
	RecoveryPhysicalPower int    `json:"recovery_physical_power"` //恢复体力
	ReclamationTime       int    `json:"reclamation_time"`        //屯田消耗时间，单位秒
	ReclamationCost       int    `json:"reclamation_cost"`        //屯田消耗政令
	DrawGeneralCost       int    `json:"draw_general_cost"`       //抽卡消耗金币
	PrPoint               int    `json:"pr_point"`                //合成一个武将或者的技能点
	Limit                 int    `json:"limit"`                   //武将数量上限

}

type role struct {
	Des               string `json:"des"`
	Wood              int    `json:"wood"`
	Iron              int    `json:"iron"`
	Stone             int    `json:"stone"`
	Grain             int    `json:"grain"`
	Gold              int    `json:"gold"`
	Decree            int    `json:"decree"`
	WoodYield         int    `json:"wood_yield"`
	IronYield         int    `json:"iron_yield"`
	StoneYield        int    `json:"stone_yield"`
	GrainYield        int    `json:"grain_yield"`
	GoldYield         int    `json:"gold_yield"`
	DepotCapacity     int    `json:"depot_capacity"` //仓库初始容量
	BuildLimit        int    `json:"build_limit"`    //野外建筑上限
	RecoveryTime      int    `json:"recovery_time"`
	DecreeLimit       int    `json:"decree_limit"`        //令牌上限
	CollectTimesLimit int8   `json:"collect_times_limit"` //每日征收次数上限
	CollectInterval   int    `json:"collect_interval"`    //征收间隔
	PosTagLimit       int8   `json:"pos_tag_limit"`       //位置标签上限
}

type city struct {
	Des           string `json:"des"`
	Cost          int8   `json:"cost"`
	Durable       int    `json:"durable"`
	RecoveryTime  int    `json:"recovery_time"`
	TransformRate int    `json:"transform_rate"`
}

type build struct {
	Des           string `json:"des"`
	WarFree       int64  `json:"war_free"`       //免战时间，单位秒
	GiveUpTime    int64  `json:"giveUp_time"`    //建筑放弃时间
	FortressLimit int    `json:"fortress_limit"` //要塞上限
}

type npcLevel struct {
	Soilders int `json:"soilders"`
}

type npc struct {
	Des    string     `json:"des"`
	Levels []npcLevel `json:"levels"`
}

type union struct {
	Des         string `json:"des"`
	MemberLimit int    `json:"member_limit"`
}
