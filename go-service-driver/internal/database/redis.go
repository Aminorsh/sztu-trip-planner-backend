package database

import (
	"log"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background()
)

func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		return err
	}

	log.Println("Redis连接成功")
	return nil
}
