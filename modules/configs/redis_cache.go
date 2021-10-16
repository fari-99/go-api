package configs

import (
	"os"
	"strconv"
	"sync"

	"github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"github.com/vmihailenco/msgpack"
)

type RedisCacheConfig struct {
	RedisCache *cache.Codec
}

var redisCacheInstance *RedisCacheConfig
var redisCacheOnce sync.Once

func GetRedisCache() *cache.Codec {
	redisCacheOnce.Do(func() {
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

		redisCacheInstance = &RedisCacheConfig{
			RedisCache: redisCache,
		}
	})

	return redisCacheInstance.RedisCache
}
