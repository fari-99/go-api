package configs

import (
	"go-api/modules/configs/rabbitmq"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/cache/v9"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/redis/go-redis/v9"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

type DI struct {
	DB            *gorm.DB
	Redis         *redis.Client
	RedisCache    *cache.Cache
	Queue         *rabbitmq.QueueSetup
	EmailDialler  *gomail.Dialer
	ElasticSearch *elasticsearch.Client
	Telegram      *tgbotapi.BotAPI
}

func DIInit() *DI {
	di := &DI{
		DB:            DatabaseBase(MySQLType).GetMysqlConnection(true),
		ElasticSearch: GetElasticSearch(),
		EmailDialler:  GetEmail(),
		Redis:         GetRedisSessionConfig(),
		RedisCache:    GetRedisCache(),
		Queue:         rabbitmq.NewBaseQueue("", ""),
		// Telegram:      GetTelegram(), // TODO: create new bot, old one deprecated
	}

	return di
}
