package database

import (
	"log"

	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	dsn := config.GetDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库连接错误：" + err.Error())
	}
	log.Println("数据库连接成功")
	DB = db
	return nil
}
