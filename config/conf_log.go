package config

type LogServer struct {
	File_die    string `yaml:"file_die"`
	Maxsize     int    `yaml:"maxsize"`
	Max_backups int    `yaml:"max_backups"`
	Max_age     int    `yaml:"max_age"`
	Compress    bool   `yaml:"compress"`
}
