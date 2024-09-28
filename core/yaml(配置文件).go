package core

import (
	"Three_kingdoms_SLG/config"
	"Three_kingdoms_SLG/global"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
)

const ConfigFile = "/conf/conf.yaml"

// yaml相关操作

// read the configuration of the yaml files
func InitConf() {
	c := &config.Config{}
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	configPath := currentDir + ConfigFile
	if !fileExist(configPath) {
		panic("yaml is not exist")
	}
	yamlConf, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(fmt.Errorf("get yamlconf error: %s", err))
	}
	err = yaml.Unmarshal(yamlConf, c)
	if err != nil {
		log.Fatalf("config Init Unmarshal: %v", err)
	}
	log.Println("config yamlFile load Init success.")
	global.Config = c
}
func fileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil || os.IsExist(err)
}

//func SetYaml() error {
//	//yaml.Unmarshal()//转为结构体
//	byteData, err := yaml.Marshal(global.Config) //转为byte
//	//将全局变量 global.Config 序列化为 YAML 格式的字节数组 byteData。
//	if err != nil {
//		global.Log.Error(err)
//		return err
//	}
//
//	err = ioutil.WriteFile(ConfigFile, byteData, fs.ModePerm) //将字节数组 byteData 写入 setting.yaml 文件。
//	if err != nil {
//		global.Log.Error(err)
//		return err
//	}
//	global.Log.Info("配置文件修改成功")
//	return nil
//}
