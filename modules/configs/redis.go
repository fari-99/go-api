package configs

import (
	"fmt"
	"github.com/go-redis/redis"
	"os"
	"strconv"
	"sync"
	"time"
)

type redisSessionConfig struct {
	session *redis.Client
}

var redisSessionInstance *redisSessionConfig
var redisSessionOnce sync.Once

func GetRedisSessionConfig() *redis.Client {
	redisSessionOnce.Do(func() {
		database, _ := strconv.Atoi(os.Getenv("REDIS_SESSION_DB"))
		timeout, _ := strconv.Atoi(os.Getenv("REDIS_SESSION_TIMEOUT"))
		minIdleConnection, _ := strconv.Atoi(os.Getenv("REDIS_SESSION_MIN_IDLE"))

		redisApp := redis.NewClient(&redis.Options{
			Network:      os.Getenv("REDIS_SESSION_NETWORK"),
			Addr:         fmt.Sprintf("%s:%s", os.Getenv("REDIS_SESSION_HOST"), os.Getenv("REDIS_SESSION_PORT")),
			Password:     os.Getenv("REDIS_SESSION_PASSWORD"),
			DB:           database,
			MaxRetries:   3,
			DialTimeout:  time.Duration(timeout) * time.Second,
			ReadTimeout:  time.Duration(timeout) * time.Second,
			WriteTimeout: time.Duration(timeout) * time.Second,
			MinIdleConns: minIdleConnection,
			MaxConnAge:   24 * time.Hour,
			TLSConfig:    nil,
		})

		_, err := redisApp.Ping().Result()
		if err != nil {
			panic(err.Error())
		}

		redisSessionInstance = &redisSessionConfig{
			session: redisApp,
		}
	})

	return redisSessionInstance.session
}
