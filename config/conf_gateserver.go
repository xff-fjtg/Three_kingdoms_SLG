package config

type GateServer struct {
	Host        string `yaml:"host"`
	Port        string `yaml:"port"`
	LoginProxy  string `yaml:"login_proxy"`
	GameProxy   string `yaml:"game_proxy"`
	Need_secret bool   `yaml:"need_secret"`
	ChatProxy   string `yaml:"chat_proxy"`
}
