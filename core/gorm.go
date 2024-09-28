package core

import (
	"Three_kingdoms_SLG/global"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"time"
)

func InitGorm() *gorm.DB {
	if global.Config.Mysql.Host == " " {
		log.Println("mysql not configured,cancel link")
		return nil
	}
	dsn := global.Config.Mysql.Dsn()
	var mysqlLogger logger.Interface
	mysqlLogger = logger.Default.LogMode(logger.Error)
	//global.MysqlLog = logger.Default.LogMode(logger.Info)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: mysqlLogger,
	})
	if err != nil {
		log.Fatalf(fmt.Sprintf("[%s] mysql fail to connection", dsn))
	}
	sqlDb, _ := db.DB()
	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetMaxOpenConns(100)
	sqlDb.SetConnMaxLifetime(time.Hour * 4)
	log.Println("connect mysql success")
	return db
}
