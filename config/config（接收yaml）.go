package config

type Config struct {
	Login      Login      `yaml:"Login"`
	Mysql      Mysql      `yaml:"mysql"`
	WebServer  WebServer  `yaml:"webserver"`
	GateServer GateServer `yaml:"gate_server"`
	GameServer GameServer `yaml:"game_server"`
	ChatServer ChatServer `yaml:"chat_server"`
	LogServer  LogServer  `yaml:"log_server"`
}
