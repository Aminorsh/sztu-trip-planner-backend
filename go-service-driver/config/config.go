package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`
	JWT struct {
		Key string `yaml:"key"`
	} `yaml:"jwt"`
	Cors struct {
		AllowedOrigins []string `yaml:"allowed_origins"`
	} `yaml:"cors"`
}

var AppConfig *Config

func LoadConfig() *Config {
	var cfg Config

	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	if err := godotenv.Load(); err != nil {
		if dsn := os.Getenv("DB_DSN"); dsn != "" {
			cfg.Database.DSN = dsn
		}
		if key := os.Getenv("JWT_KEY"); key != "" {
			cfg.JWT.Key = key
		}
	}

	AppConfig = &cfg
	return &cfg
}

func GetDSN() string {
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DB_NAME")

	return user + ":" + pass + "@tcp(" + host + ":" + port + ")/" + name + "?charset=utf8mb4&parseTime=True&loc=Local"
}

func GetJWTKey() []byte {
	return []byte(os.Getenv("JWT_KEY"))
}

func GetMailAPIKey() string {
	return os.Getenv("MAIL_API_KEY")
}
