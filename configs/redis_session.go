package configs

import (
	"os"
	"strconv"
	"sync"

	"github.com/go-redis/redis"
)

type RedisSessionConfig struct {
	Client *redis.Client
}

var redisSessionInstance *RedisSessionConfig
var redisSessionOnce sync.Once

func GetRedisSession() *redis.Client {
	redisSessionOnce.Do(func() {
		redisDB, _ := strconv.ParseInt(os.Getenv("REDIS_SESSION_DB"), 10, 64)

		client := redis.NewClient(&redis.Options{
			Addr: os.Getenv("REDIS_SESSION_HOST") + ":" + os.Getenv("REDIS_SESSION_PORT"),
			//Password: os.Getenv("REDIS_SESSION_PASSWORD"),
			DB: int(redisDB),
		})

		client.TTL(os.Getenv("REDIS_SESSION_LIFETIME"))

		redisSessionInstance = &RedisSessionConfig{
			Client: client,
		}
	})

	return redisSessionInstance.Client
}
