package controllers

import (
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

type ArticlesController struct {
	DB    *gorm.DB
	Redis *redis.Client
}
