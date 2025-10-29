package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func loadConfig() {
	// 配置加载逻辑
	err := godotenv.Load()
	if err != nil {
		log.Println("无法加载 .env 文件，使用默认配置")
	}
}

func GetDSN() string {
	loadConfig()
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	return user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + name + "?charset=utf8mb4&parseTime=True&loc=Local"
}

func GetJWTKey() []byte {
	loadConfig()
	return []byte(os.Getenv("JWT_KEY"))
}

func GetServerPort() string {
	loadConfig()
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080" // 默认端口
	}
	return port
}
