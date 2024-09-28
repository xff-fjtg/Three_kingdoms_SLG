package global

import (
	"Three_kingdoms_SLG/config"
	"gorm.io/gorm"
)

var (
	Config *config.Config
	DB     *gorm.DB
)
