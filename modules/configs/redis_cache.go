package configs

import (
	"log"
	"os"
	"strconv"

	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/vmihailenco/msgpack/v5"
)

func GetRedisCache() *cache.Cache {
	log.Println("Initialize Redis Cache connection...")

	redisCacheDB, _ := strconv.ParseInt(os.Getenv("REDIS_CACHE_DB"), 10, 64)

	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_CACHE_HOST") + ":" + os.Getenv("REDIS_CACHE_PORT"),
		// Password: os.Getenv("REDIS_CACHE_PASSWORD"),
		DB: int(redisCacheDB),
	})

	redisCache := cache.New(&cache.Options{
		Redis: client,
		Marshal: func(i interface{}) (bytes []byte, e error) {
			return msgpack.Marshal(i)
		},
		Unmarshal: func(bytes []byte, i interface{}) error {
			return msgpack.Unmarshal(bytes, i)
		},
	})

	log.Println("Success Initialize Redis Cache connection...")
	return redisCache
}
