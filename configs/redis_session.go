package configs

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"github.com/kataras/iris/v12/sessions/sessiondb/redis"
	"os"
	"strconv"
	"sync"
	"time"
)

type redisSessionConfig struct {
	session *sessions.Sessions
}

var redisSessionInstance *redisSessionConfig
var redisSessionOnce sync.Once

func GetRedisSessionConfig() *sessions.Sessions {
	redisSessionOnce.Do(func() {
		timeout, _ := strconv.ParseInt(os.Getenv("REDIS_SESSION_TIMEOUT"), 10, 64)
		maxIdle, _ := strconv.ParseInt(os.Getenv("REDIS_SESSION_MAX_IDLE"), 10, 64)
		timeoutRedis := time.Duration(timeout) * time.Minute

		redisDriver := redis.Redigo()
		redisDriver.MaxIdle = int(maxIdle)
		redisDriver.IdleTimeout = timeoutRedis

		sessionDB := redis.New(redis.Config{
			Network:   os.Getenv("REDIS_SESSION_NETWORK"),
			Addr:      fmt.Sprintf("%s:%s", os.Getenv("REDIS_SESSION_HOST"), os.Getenv("REDIS_SESSION_PORT")),
			Database:  os.Getenv("REDIS_SESSION_DB"),
			MaxActive: 0,
			Timeout:   timeoutRedis,
			Driver:    redisDriver,
			//Clusters:  nil,
			//Password:  "",
			//Prefix:    "",
			//Delim:     "",
		})

		iris.RegisterOnInterrupt(func() {
			sessionDB.Close()
		})

		timeDuration, _ := strconv.ParseInt(os.Getenv("REDIS_SESSION_EXPIRED"), 10, 64)
		sessionRedisAccess := sessions.New(sessions.Config{
			Cookie:  fmt.Sprintf("_session_id"),
			Expires: time.Duration(timeDuration) * time.Hour,
		})

		redisSessionInstance = &redisSessionConfig{
			session: sessionRedisAccess,
		}
	})

	return redisSessionInstance.session
}
