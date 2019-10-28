package configs

import (
	"os"
	"strconv"
	"sync"

	"github.com/go-redis/redis"
)

type RedisConfig struct {
	Client *redis.Client
}

var redisSessionInstance *RedisConfig
var redisOnce sync.Once

func GetRedis() *redis.Client {
	redisOnce.Do(func() {
		redisDB, _ := strconv.ParseInt(os.Getenv("REDIS_SESSION_DB"), 10, 64)

		client := redis.NewClient(&redis.Options{
			Addr: os.Getenv("REDIS_SESSION_HOST") + ":" + os.Getenv("REDIS_SESSION_PORT"),
			//Password: os.Getenv("REDIS_SESSION_PASSWORD"),
			DB: int(redisDB),
		})

		client.TTL(os.Getenv("REDIS_SESSION_LIFETIME"))

		redisSessionInstance = &RedisConfig{
			Client: client,
		}
	})

	return redisSessionInstance.Client
}
