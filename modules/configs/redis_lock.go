package configs

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/spf13/cast"
)

func getRedisLockConfig() redis.Options {
	redisLockHost := os.Getenv("REDIS_LOCK_HOST")
	redisLockPort := os.Getenv("REDIS_LOCK_PORT")
	redisLockNetwork := os.Getenv("REDIS_LOCK_NETWORK")
	redisLockDB := os.Getenv("REDIS_LOCK_DB")
	redisLockUsername := os.Getenv("REDIS_LOCK_USERNAME")
	redisLockPassword := os.Getenv("REDIS_LOCK_PASSWORD")
	redisLockMaxRetry := os.Getenv("REDIS_LOCK_MAX_RETRY")
	redisLockMinIdleConn := os.Getenv("REDIS_LOCK_MIN_IDLE_CONN")

	redisOption := redis.Options{
		Network:      redisLockNetwork,
		Addr:         fmt.Sprintf("%s:%s", redisLockHost, redisLockPort),
		DB:           cast.ToInt(redisLockDB),
		MaxRetries:   cast.ToInt(redisLockMaxRetry),
		PoolFIFO:     true,
		MinIdleConns: cast.ToInt(redisLockMinIdleConn),
	}

	if redisLockUsername != "" && redisLockPassword != "" {
		redisOption.Username = redisLockUsername
		redisOption.Password = redisLockPassword
	}

	return redisOption
}

func GetRedisLock() *redsync.Redsync {
	redisOption := getRedisLockConfig()
	client := redis.NewClient(&redisOption)

	err := client.Ping(context.Background()).Err()
	if err != nil {
		panic(err.Error())
	}

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	log.Printf("Success Connecting to redis lock")
	return rs
}
