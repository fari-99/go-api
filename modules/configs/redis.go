package configs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const REDIS_SESSION_PREFIX = "REDIS_SESSION"
const REDIS_COUNTING_PREFIX = "REDIS_COUNTING"

// RedisInstance holds a named redis client
type RedisInstance struct {
	client redis.UniversalClient
}

var (
	redisInstances = make(map[string]*RedisInstance)
	redisMu        sync.Mutex
	redisOnce      = make(map[string]*sync.Once)
)

// RedisConfig holds the configuration for a RedisSession connection
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	Timeout  time.Duration
	MinIdle  int
}

// redisEnvConfig builds a RedisConfig from env variables by prefix.
// e.g. prefix "REDIS_SESSION" reads REDIS_SESSION_HOST, REDIS_SESSION_PORT, etc.
func redisEnvConfig(prefix string) (RedisConfig, error) {
	db, _ := strconv.Atoi(os.Getenv(prefix + "_DB"))
	timeout, _ := strconv.Atoi(os.Getenv(prefix + "_TIMEOUT"))
	minIdle, _ := strconv.Atoi(os.Getenv(prefix + "_MIN_IDLE"))
	host := os.Getenv(prefix + "_HOST")
	port := os.Getenv(prefix + "_PORT")

	if host == "" || port == "" {
		return RedisConfig{}, errors.New("host or port is empty for prefix " + prefix)
	}

	return RedisConfig{
		Host:     host,
		Port:     port,
		Password: os.Getenv(prefix + "_PASSWORD"),
		DB:       db,
		Timeout:  time.Duration(timeout) * time.Second,
		MinIdle:  minIdle,
	}, nil
}

// GetRedis returns a singleton RedisSession client by prefixes.
// If it doesn't exist yet, it initializes it using the provided config.
func GetRedis(prefix string) redis.UniversalClient {
	cfg, err := redisEnvConfig(prefix)
	if err != nil {
		panic(err)
	}

	redisMu.Lock()
	if _, ok := redisOnce[prefix]; !ok {
		redisOnce[prefix] = &sync.Once{}
	}
	once := redisOnce[prefix]
	redisMu.Unlock()

	once.Do(func() {
		log.Printf("Initializing RedisSession [%s] connection...\n", prefix)

		client := redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:        []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
			Password:     cfg.Password,
			DB:           cfg.DB,
			MaxRetries:   3,
			DialTimeout:  cfg.Timeout,
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			MinIdleConns: cfg.MinIdle,
		})

		if _, err := client.Ping(context.Background()).Result(); err != nil {
			panic(fmt.Sprintf("RedisSession [%s] connection failed: %s", prefix, err.Error()))
		}

		redisMu.Lock()
		redisInstances[prefix] = &RedisInstance{client: client}
		redisMu.Unlock()

		log.Printf("RedisSession [%s] connected successfully.\n", prefix)
	})

	redisMu.Lock()
	defer redisMu.Unlock()
	return redisInstances[prefix].client
}
