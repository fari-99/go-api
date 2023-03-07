package configs

import (
	"os"
	"strconv"

	"github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"github.com/vmihailenco/msgpack"
)

func GetRedisCache() *cache.Codec {
	redisCacheDB, _ := strconv.ParseInt(os.Getenv("REDIS_CACHE_DB"), 10, 64)

	client := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_CACHE_HOST") + ":" + os.Getenv("REDIS_CACHE_PORT"),
		//Password: os.Getenv("REDIS_CACHE_PASSWORD"),
		DB: int(redisCacheDB),
	})

	redisCache := &cache.Codec{
		Redis: client,
		Marshal: func(i interface{}) (bytes []byte, e error) {
			return msgpack.Marshal(i)
		},
		Unmarshal: func(bytes []byte, i interface{}) error {
			return msgpack.Unmarshal(bytes, i)
		},
	}

	return redisCache
}
