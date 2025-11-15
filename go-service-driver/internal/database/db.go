package database

import (
	"log"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	dsn := config.GetDSN()
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	// if err != nil {
	// 	panic("数据库连接错误：" + err.Error())
	// }
	// log.Println("数据库连接成功")
	// DB = db
	// return nil

	for i := range 10 {
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Printf("数据库连接错误，第 %d 次重试: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		} else {
			log.Println("数据库连接成功")
			DB = db
			return nil
		}
	}

	panic("数据库连接失败，已达最大重试次数")
}
