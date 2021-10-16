package configs

import (
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/cache"
	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	"gopkg.in/gomail.v2"
)

type DI struct {
	DB            *gorm.DB
	Redis         *redis.Client
	RedisCache    *cache.Codec
	Queue         *QueueSetup
	EmailDialler  *gomail.Dialer
	ElasticSearch *elasticsearch.Client
	Telegram      *tgbotapi.BotAPI
}

func DIInit() *DI {
	di := &DI{
		DB:            DatabaseBase().GetDBConnection(),
		ElasticSearch: GetElasticSearch(),
		EmailDialler:  GetEmail(),
		Queue:         NewBaseQueue().GetQueueUtil(),
		Redis:         GetRedisSessionConfig(),
		RedisCache:    GetRedisCache(),
		Telegram:      GetTelegram(),
	}

	return di
}
