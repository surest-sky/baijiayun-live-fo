package request

import (
	"github.com/go-redis/redis"
	"talk/config"
	"talk/global"
)

func RedisClient() {
	client := redis.NewClient(&redis.Options{
		Addr:     config.GetString("REDIS_HOST") + ":6379",
		Password: "AdAd123AdAd",
		DB:       5,
	})

	_, err := client.Ping().Result()
	global.PanicError(err, "redis 连接")
	global.REDIS_CLIENT = client
}
