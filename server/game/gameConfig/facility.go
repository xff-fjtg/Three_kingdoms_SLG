package gameConfig

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const (
	TypeDurable        = 1 //耐久
	TypeCost           = 2
	TypeArmyTeams      = 3 //队伍数量
	TypeSpeed          = 4 //速度
	TypeDefense        = 5 //防御
	TypeStrategy       = 6 //谋略
	TypeForce          = 7 //攻击武力
	TypeConscriptTime  = 8 //征兵时间
	TypeReserveLimit   = 9 //预备役上限
	TypeUnkonw         = 10
	TypeHanAddition    = 11
	TypeQunAddition    = 12
	TypeWeiAddition    = 13
	TypeShuAddition    = 14
	TypeWuAddition     = 15
	TypeDealTaxRate    = 16 //交易税率
	TypeWood           = 17
	TypeIron           = 18
	TypeGrain          = 19
	TypeStone          = 20
	TypeTax            = 21 //税收
	TypeExtendTimes    = 22 //扩建次数
	TypeWarehouseLimit = 23 //仓库容量
	TypeSoldierLimit   = 24 //带兵数量
	TypeVanguardLimit  = 25 //前锋数量
)

type conditions struct {
	Type  int `json:"type"`
	Level int `json:"level"`
}

type facilityConf struct { //对应facility json里面的数据
	Title     string `json:"title"`
	List      []conf `json:"list"`
	facilitys map[int8]*facility
}
type facility struct {
	Title      string       `json:"title"`
	Des        string       `json:"des"`
	Name       string       `json:"name"`
	Type       int8         `json:"type"`
	Additions  []int8       `json:"additions"`
	Conditions []conditions `json:"conditions"`
	Levels     []fLevel     `json:"levels"`
}

type NeedRes struct {
	Decree int `json:"decree"`
	Grain  int `json:"grain"`
	Wood   int `json:"wood"`
	Iron   int `json:"iron"`
	Stone  int `json:"stone"`
	Gold   int `json:"gold"`
}

type fLevel struct {
	Level  int     `json:"level"`
	Values []int   `json:"values"`
	Need   NeedRes `json:"need"`
	Time   int     `json:"time"` //升级需要的时间
}

type conf struct {
	Name string
	Type int8
}

var FacilityConf = &facilityConf{}

const facilityFile = "/conf/facility/facility.json"
const facilityPath = "/conf/facility/"

func (f *facilityConf) Load() {
	//获取当前文件路径
	currentDir, _ := os.Getwd()
	//配置文件位置
	configFile := currentDir + facilityFile
	confPath := currentDir + facilityPath
	//打包后 程序参数加入配置文件路径
	if len(os.Args) > 1 {
		if path := os.Args[1]; path != "" {
			configFile = path + facilityFile
			confPath = path + facilityPath
		}
	}
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("facility读取失败")
		panic(err)
	}
	err = json.Unmarshal(data, f)
	if err != nil {
		log.Println("facility json 转换失败")
		panic(err)
	}
	f.facilitys = make(map[int8]*facility, len(f.List))
	files, err := ioutil.ReadDir(confPath)
	if err != nil {
		log.Println("facility confPath 读取失败")
		panic(err)
	}
	for _, files := range files {
		if files.IsDir() {
			continue
		}
		fileName := files.Name()
		if fileName == "facility.json" {
			continue
		}

		filePath := confPath + fileName
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Println("facility confPath 读取失败")
			panic(err)
		}
		fac := &facility{}
		err = json.Unmarshal(data, fac)
		if err != nil {
			log.Println("facility data 转换失败")
		}
		f.facilitys[fac.Type] = fac
	}
}

// 下面代码可以对应facility_produce_mj.json看
func (f *facilityConf) CostTime(fType int8, level int8) int {
	if level <= 0 {
		return 0
	}
	fa, ok := f.facilitys[fType]
	if ok {
		if int8(len(fa.Levels)) >= level {
			return fa.Levels[level-1].Time - 2 //比客户端快2s，保证客户端倒计时完一定是升级成功了
		} else {
			return 0
		}
	} else {
		return 0
	}
}

func (f *facilityConf) GetValues(fType int8, level int8) []int {
	if level <= 0 {
		return []int{}
	}

	fa, ok := f.facilitys[fType]
	if ok {
		if int8(len(fa.Levels)) >= level {
			return fa.Levels[level-1].Values
		} else {
			return []int{}
		}
	} else {
		return []int{}
	}
}

func (f *facilityConf) GetAdditions(fType int8) []int8 {
	fa, ok := f.facilitys[fType]
	if ok {
		return fa.Additions
	} else {
		return []int8{}
	}
}

func (f *facilityConf) MaxLevel(fType int8) int8 {
	fac, ok := f.facilitys[fType]
	if ok {
		return int8(len(fac.Levels))
	} else {
		return 0
	}
}

func (f *facilityConf) Need(fType int8, level int8) (*NeedRes, bool) {
	if level <= 0 {
		return nil, false
	}

	fac, ok := f.facilitys[fType]
	if ok {
		if int8(len(fac.Levels)) >= level {
			return &fac.Levels[level-1].Need, true
		} else {
			return nil, false
		}
	} else {
		return nil, false
	}

}

const (
	Main          = 0  //主城
	JiaoChang     = 13 //校场
	TongShuaiTing = 14 //统帅厅
	JiShi         = 15 //集市
	MBS           = 16 //募兵所
)
