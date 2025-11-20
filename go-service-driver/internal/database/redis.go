package database

import (
	"log"
	"time"

	"github.com/Aminorsh/sztu-trip-planner-backend/config"
	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background()
)

func InitRedis() error {
	// RedisClient = redis.NewClient(&redis.Options{
	// 	Addr:     config.GetRedisAddr(),
	// 	Password: "", // no password set
	// 	DB:       0,  // use default DB
	// })

	// _, err := RedisClient.Ping(Ctx).Result()
	// if err != nil {
	// 	return err
	// }

	// log.Println("Redis连接成功")
	// return nil
	for i := range 10 {
		RedisClient = redis.NewClient(&redis.Options{
			Addr:     config.GetRedisAddr(),
			Password: "", // no password set
			DB:       0,  // use default DB
		})

		_, err := RedisClient.Ping(Ctx).Result()
		if err != nil {
			log.Printf("Redis连接错误，第 %d 次重试: %v", i+1, err)
			time.Sleep(2 * time.Second)
			continue
		} else {
			log.Println("Redis连接成功")
			return nil
		}
	}

	panic("Redis连接失败，已达最大重试次数")
}
